package runner

import (
	"context"
	"github.com/beetbasket/runner/pkg/matcher"
	"github.com/beetbasket/runner/pkg/message"
	"github.com/beetbasket/runner/pkg/message/input"
	"github.com/beetbasket/runner/pkg/message/output"
	"github.com/beetbasket/runner/pkg/stdionet"
	"github.com/beetbasket/rx"
	"github.com/point-c/wg"
	"io"
	"slices"
)

func (cmd *Cmd) newKindWriters() (*kindWriter[output.StdoutMessage], *kindWriter[output.StderrMessage]) {
	return &kindWriter[output.StdoutMessage]{
			out:      &cmd.out,
			ctx:      cmd.ctx,
			matcher:  matcher.New(cmd.prefix),
			netstack: cmd.netstack,
		}, &kindWriter[output.StderrMessage]{
			out:      &cmd.out,
			ctx:      cmd.ctx,
			matcher:  matcher.New(""),
			netstack: cmd.netstack,
		}
}

type kindWriter[K output.StdioLike] struct {
	out      *rx.Subject[message.Message]
	netstack *wg.Netstack
	ctx      context.Context
	matcher  *matcher.Matcher
}

func (kw *kindWriter[K]) Write(b []byte) (n int, _ error) {
	if kw.ctx.Err() != nil {
		return 0, kw.ctx.Err()
	}

	n, _ = kw.matcher.Write(b)
	if b := kw.matcher.ReadOut(); len(b) > 0 {
		kw.out.Next(output.NewStdioMessage[K](b))
	}

	if b := kw.matcher.ReadSpecial(); len(b) > 0 {
		stdionet.DecodePackets(kw.netstack, b)
	}
	return len(b), nil
}

func (cmd *Cmd) pipeInput(in io.WriteCloser) {
	defer in.Close()
	defer cmd.cancel()

	ctx, cancel := context.WithCancel(cmd.ctx)
	defer cancel()
	stdin := cmd.in.Subscribe(ctx)
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-stdin:
			if ok {
				b := data.Input()
				if _, err := in.Write(b); err != nil {
					return
				} else if _, ok := data.(input.PacketInput); !ok {
					cmd.out.Next(output.NewStdioMessage[output.StdinMessage](b))
				}
			} else {
				return
			}
		}
	}
}

func (cmd *Cmd) pipePackets() {
	defer cmd.cancel()
	buf := [...][]byte{make([]byte, wg.DefaultMTU*2)}
	var size [len(buf)]int
	for cmd.ctx.Err() == nil {
		n, err := cmd.netstack.Read(buf[:], size[:], 0)
		if err != nil {
			return
		} else if n > 0 {
			for i, b := range buf[:n] {
				if size[i] > 0 {
					cmd.in.Next(input.NewPacketInput(cmd.prefix, slices.Clone(b[:size[i]])))
				}
			}
		}
	}
}
