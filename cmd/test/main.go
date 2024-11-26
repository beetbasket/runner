package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/beetbasket/runner"
	"github.com/beetbasket/runner/pkg/message/output"
	"github.com/beetbasket/runner/pkg/stdionet"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "child" {
		childMain()
	} else {
		parentMain()
	}
}

func parentMain() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := must(runner.New(ctx, runner.NewCommandArgs(must(os.Executable()), []string{"child"})))
	defer do(cmd.Close)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)).With(slog.Bool("parent", true)))

	var wg sync.WaitGroup
	wg.Add(1)
	oc := cmd.Output(ctx)
	go func() {
		var buf bytes.Buffer
		defer wg.Done()
		slog.Info("listening to child output")
		for out := range oc {
			switch out := out.(type) {
			case output.StartMessage:
				slog.Info("child started")
			case output.ExitMessage:
				return
			case output.StderrMessage:
				os.Stderr.Write(out.Data)
			case output.StdoutMessage:
				os.Stdout.Write(out.Data)
			default:
				buf.Reset()
				check(json.NewEncoder(&buf).Encode(out))
				slog.Info("got other output from child", slog.String("out", buf.String()))
			}
		}
	}()

	ln := must(cmd.Listen(80))
	defer do(ln.Close)
	go func() {
		defer cancel()
		slog.Error("http server exited", slog.Any("error", http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var msg Message
			check(json.NewDecoder(r.Body).Decode(&msg))
			ipv4 := net.IPv4(msg.IP[0], msg.IP[1], msg.IP[2], msg.IP[3])
			slog.Info("got http from child", slog.String("ip", ipv4.String()), slog.Any("port", msg.Port), slog.String("data", msg.Data))
			json.NewEncoder(w).Encode(Message{
				IP:   [4]byte{1, 2, 3, 4},
				Port: 1234,
				Data: "hello client",
			})
		}))))
	}()

	cmd.Start()
	select {
	case <-ctx.Done():
	case <-cmd.Wait():
		wg.Wait()
	}
}

type Message struct {
	IP   [4]byte
	Port uint16
	Data string
}

func childMain() {
	slog.SetDefault(slog.New(slog.NewTextHandler(stdionet.Environment().Stdout(), nil)).With(
		slog.Bool("child", true),
	))

	slog.Info("child started")
	nt := stdionet.Environment()

	resp := must((&http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				parentAddr := nt.ParentAddrTCP(80)
				slog.Info("child dialing", slog.Any("parent", parentAddr))
				return nt.Dial(context.Background(), parentAddr)
			},
		},
	}).Post("http://localhost:80", "application/json", func() io.Reader {
		var buf bytes.Buffer
		check(json.NewEncoder(&buf).Encode(Message{
			IP:   [4]byte{nt.Address()[0], nt.Address()[1], nt.Address()[2], nt.Address()[3]},
			Port: 80,
			Data: "hello server",
		}))
		slog.Info("child making request", slog.String("data", buf.String()))
		return &buf
	}()))
	must(io.Copy(os.Stdout, resp.Body))
	resp.Body.Close()
	slog.Info("request made, exiting")
	nt.Exit(5)
}

func must[T any](v T, err error) T {
	check(err)
	return v
}

func do(fn func() error) {
	check(fn())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
