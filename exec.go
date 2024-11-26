package runner

import (
	"context"
	"fmt"
	"github.com/beetbasket/runner/pkg/message"
	"github.com/beetbasket/runner/pkg/message/input"
	"github.com/beetbasket/runner/pkg/message/output"
	"github.com/beetbasket/runner/pkg/stdionet"
	"github.com/beetbasket/rx"
	"github.com/google/uuid"
	"github.com/point-c/wg"
	"github.com/trymoose/errors"
	"io"
	"math/rand/v2"
	"net"
	"os/exec"
	"strings"
	"sync/atomic"
)

type Cmd struct {
	in  rx.Subject[message.Input]
	out rx.Subject[message.Message]

	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc

	netstack *wg.Netstack
	prefix   string
	address  net.IP

	started atomic.Bool
	wait    chan struct{}
	waitErr error
}

func New(ctx context.Context, cmd CommandArgsEnv) (_ *Cmd, finalErr error) {
	finally, cleanup := CheckOk()
	// Setup networking
	netstack, err := wg.NewDefaultNetstack()
	if err != nil {
		return nil, err
	}
	defer cleanup(func() { finalErr = errors.Join(finalErr, netstack.Close()) })

	// Setup command struct
	ctx, cancel := context.WithCancel(ctx)
	defer cleanup(cancel)
	c := Cmd{
		ctx:      ctx,
		cancel:   cancel,
		netstack: netstack,
		prefix:   generatePrefix(),
		address:  stdionet.GenerateRandomIPv4(),
		wait:     make(chan struct{}),
	}

	// Make command and setup io
	in, err := c.initializeCommand(cmd)
	if err != nil {
		return nil, err
	}

	// Copy io goroutines
	// Make sure close is run at lease once if one of the goroutines cancels the context
	stop := context.AfterFunc(ctx, func() { c.Close() })
	defer cleanup(func() { stop() })
	go c.pipeInput(in)
	go c.pipePackets()

	finally()
	return &c, nil
}

func CheckOk() (finally func(), cleanup func(func())) {
	ok := false
	return func() {
			ok = true
		}, func(fn func()) {
			if !ok {
				fn()
			}
		}
}

func (cmd *Cmd) Input(in message.Input) {
	if _, ok := in.(input.PacketInput); ok || in == nil {
		return
	}
	cmd.in.Next(in)
}

func (cmd *Cmd) Dial(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	return cmd.netstack.Net().Dialer(cmd.address, 0).DialTCP(ctx, addr)
}

func (cmd *Cmd) Listen(port uint16) (net.Listener, error) {
	return cmd.netstack.Net().Listen(&net.TCPAddr{
		IP:   cmd.address,
		Port: int(port),
	})
}

func (cmd *Cmd) Output(ctx context.Context) <-chan message.Message {
	return cmd.out.Subscribe(ctx)
}

func (cmd *Cmd) Start() {
	if cmd.started.CompareAndSwap(false, true) {
		go cmd.runCmd()
	}
}

func (cmd *Cmd) runCmd() {
	defer cmd.cleanupCmd(true)
	cmd.out.Next(output.NewStartMessage())

	if err := cmd.cmd.Run(); err != nil {
		if exit := new(exec.ExitError); errors.As(err, &exit) {
			cmd.exitComplete(exit.ExitCode())
		} else {
			cmd.exitComplete(-1)
			cmd.waitErr = errors.Join(cmd.waitErr, err)
		}
	}
}

func (cmd *Cmd) exitComplete(code int) {
	cmd.out.Complete(output.NewExitMessage(code))
}

func (cmd *Cmd) cleanupCmd(started bool) {
	cmd.waitErr = errors.Join(cmd.waitErr, cmd.netstack.Close())
	close(cmd.wait)
	if started {
		cmd.exitComplete(0)
	} else {
		cmd.out.Complete()
	}
}

func (cmd *Cmd) Wait() <-chan struct{} {
	return cmd.wait
}

func (cmd *Cmd) Close() error {
	cmd.cancel()
	if cmd.started.CompareAndSwap(false, true) {
		// never started
		cmd.cleanupCmd(false)
	} else {
		<-cmd.Wait()
	}
	return cmd.waitErr
}

const prefixChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

func generatePrefix() string {
	var sb strings.Builder
	sb.WriteString(uuid.NewString())
	sb.WriteRune('-')
	for range 10 {
		sb.WriteByte(prefixChars[rand.IntN(len(prefixChars))])
	}
	return sb.String()
}

func (cmd *Cmd) initializeCommand(cae CommandArgsEnv) (io.WriteCloser, error) {
	cmd.cmd = exec.CommandContext(cmd.ctx, cae.Command(), cae.Args()...)
	cmd.cmd.Env = append(cae.Environment(),
		fmt.Sprintf("PACKET_PREFIX=%s", cmd.prefix),
		fmt.Sprintf("PARENT_ADDRESS=%s", cmd.address.String()),
	)
	cmd.cmd.Stdout, cmd.cmd.Stderr = cmd.newKindWriters()
	return cmd.cmd.StdinPipe()
}
