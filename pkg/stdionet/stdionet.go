package stdionet

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"github.com/beetbasket/program/pkg/env"
	"github.com/beetbasket/runner/pkg/matcher"
	"github.com/beetbasket/runner/pkg/message/input"
	"github.com/point-c/wg"
	"github.com/point-c/wg/pkg/ipcheck"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

func init() {
	env.RegisterParser[IPv4Parser]()
}

type Env struct {
	Prefix  string `env:"PACKET_PREFIX" description:"Line prefix of network packets"`
	Address net.IP `env:"PARENT_ADDRESS" description:"Parent's ip address'" parser:"ipv4"`
}

type StdioNet struct {
	env     Env
	ns      *wg.Netstack
	address net.IP
	stdin   lockedBuf
}

func Environment() (*StdioNet, error) {
	sn := StdioNet{
		env:     Env{},
		ns:      must(wg.NewDefaultNetstack()),
		address: GenerateRandomIPv4(),
	}
	if err := env.Unmarshal(&sn.env); err != nil {
		return nil, err
	}
	go sn.writePackets()
	go sn.sortStdin()
	return &sn, nil
}

func (sn *StdioNet) sortStdin() {
	var buf [wg.DefaultMTU * 2]byte
	mm := matcher.New(sn.env.Prefix)
	for {
		n, err := os.Stdin.Read(buf[:])
		if err != nil {
			slog.Error("failed to read from stdin", slog.Any("error", err))
			return
		}
		mm.Write(buf[:n])

		sn.stdin.Write(mm.ReadOut())
		if b := mm.ReadSpecial(); len(b) > 0 {
			DecodePackets(sn.ns, b)
		}
	}
}

func (sn *StdioNet) writePackets() {
	buf := [...][]byte{make([]byte, wg.DefaultMTU*2)}
	var size [len(buf)]int
	for {
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
	if err := sn.ns.Close(); err != nil {
		slog.Error("error closing netstack", slog.Any("error", err))
	}

	time.Sleep(time.Second * 2)

	if len(code) > 0 {
		os.Exit(code[0])
	} else {
		os.Exit(0)
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

type IPv4Parser struct {
}

func (IPv4Parser) Name() string {
	return "ipv4"
}

func (IPv4Parser) Parse(s string) (net.IP, error) {
	return net.ParseIP(s).To4(), nil
}

func GenerateRandomIPv4() net.IP {
	var buf [4]byte
	for {
		binary.LittleEndian.PutUint32(buf[:], rand.Uint32())
		ip := net.IPv4(buf[0], buf[1], buf[2], buf[3])
		if !ipcheck.IsBogon(ip, ipcheck.IsPrivateNetwork) {
			return ip.To4()
		}
	}
}

func must[T any](t T, err error) T {
	check(err)
	return t
}

func check(err error) {
	if err != nil {
		panic(err)
	}
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

func DecodePackets(ns *wg.Netstack, b []byte) {
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		b = bytes.TrimSpace(sc.Bytes())
		if len(b) == 0 {
			continue
		}
		db, err := base64.StdEncoding.AppendDecode(nil, b)
		if err != nil {
			slog.Error("failed to decode packet data", slog.String("data", string(b)))
			continue
		}
		_, _ = ns.Write([][]byte{db}, 0)
	}
}
