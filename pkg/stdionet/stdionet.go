package stdionet

import (
	"bytes"
	"context"
	"github.com/beetbasket/program/pkg/env"
	"github.com/beetbasket/program/pkg/log"
	"github.com/beetbasket/runner/pkg/ipv4"
	"github.com/beetbasket/runner/pkg/matcher"
	"github.com/beetbasket/runner/pkg/message/input"
	"github.com/point-c/wg"
	"go.uber.org/fx"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
)

func New() fx.Option {
	return fx.Module("stdionet",
		fx.Provide(
			env.Unmarshal[Env],
			fx.Private,
		),
		fx.Provide(
			newStdionet,
		),
	)
}

type Env struct {
	Prefix  string `env:"PACKET_PREFIX" description:"Line prefix of network packets"`
	Address net.IP `env:"PARENT_ADDRESS" description:"Parent's ip address'" parser:"ipv4"`
}

type StdioNet struct {
	env      Env
	ns       *wg.Netstack
	address  net.IP
	stdin    lockedBuf
	shutdown fx.Shutdowner
}

func newStdionet(
	lf fx.Lifecycle,
	shutdown fx.Shutdowner,
	ev Env,
	ctx context.Context,
) (*StdioNet, error) {
	ns, err := wg.NewDefaultNetstack()
	if err != nil {
		return nil, err
	}
	lf.Append(fx.StopHook(func() error {
		return ns.Close()
	}))

	sn := StdioNet{
		ns:       ns,
		address:  ipv4.GenerateRandomIPv4(),
		env:      ev,
		shutdown: shutdown,
	}

	ctx, cancel := context.WithCancel(ctx)
	lf.Append(fx.StartStopHook(func() {
		go sn.writePackets(ctx)
	}, cancel))
	lf.Append(fx.StartStopHook(func() {
		go sn.sortStdin(ctx)
	}, cancel))
	return &sn, nil
}

func (sn *StdioNet) sortStdin(ctx context.Context) {
	var buf [1000]byte
	mm := matcher.New(sn.env.Prefix)
	for ctx.Err() == nil {
		n, err := os.Stdin.Read(buf[:])
		if err != nil {
			slog.Error("failed to read from stdin", slog.Any("error", err))
			return
		}
		_, _ = mm.Write(buf[:n])

		_, _ = sn.stdin.Write(mm.ReadOut())
		if b := mm.ReadSpecial(); len(b) > 0 {
			ipv4.DecodePackets(sn.ns, b)
		}
	}
}

func (sn *StdioNet) writePackets(ctx context.Context) {
	buf := [...][]byte{make([]byte, wg.DefaultMTU*2)}
	var size [len(buf)]int
	for ctx.Err() == nil {
		n, err := sn.ns.Read(buf[:], size[:], 0)
		if err != nil {
			slog.Error("error reading from netstack", slog.Any("error", err))
			return
		} else if n == 0 {
			continue
		}
		for i, b := range buf {
			data := input.NewPacketInput(sn.env.Prefix, b[:size[i]]).Input()
			if _, err := sn.Stdout().Write(data); err != nil {
				slog.Error("failed to write packet to stdout", slog.Any("error", err))
				return
			}
		}
	}
}

func (sn *StdioNet) Stdout() io.Writer {
	return os.Stdout
}

func (sn *StdioNet) Stderr() io.Writer {
	return os.Stderr
}

func (sn *StdioNet) Stdin() io.Reader {
	return &sn.stdin
}

func (sn *StdioNet) Exit(code ...int) {
	var opts []fx.ShutdownOption
	if len(code) > 0 {
		opts = append(opts, fx.ExitCode(code[0]))
	}

	if err := sn.shutdown.Shutdown(opts...); err != nil {
		slog.Error("failed to shutdown", log.Err(err))
	}
}

func (sn *StdioNet) Address() net.IP {
	return sn.address
}

func (sn *StdioNet) ParentAddr() net.IP {
	return sn.env.Address
}

func (sn *StdioNet) ParentAddrTCP(port uint16) *net.TCPAddr {
	return &net.TCPAddr{
		IP:   sn.env.Address,
		Port: int(port),
	}
}

func (sn *StdioNet) Dial(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	return sn.ns.Net().Dialer(sn.address, 0).DialTCP(ctx, addr)
}

func (sn *StdioNet) Listen(port uint16) (net.Listener, error) {
	return sn.ns.Net().Listen(&net.TCPAddr{
		IP:   sn.address,
		Port: int(port),
	})
}

type lockedBuf struct {
	buf  bytes.Buffer
	lock sync.RWMutex
}

func (lb *lockedBuf) Read(b []byte) (int, error) {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	return lb.buf.Read(b)
}

func (lb *lockedBuf) Write(b []byte) (int, error) {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	return lb.buf.Write(b)
}
