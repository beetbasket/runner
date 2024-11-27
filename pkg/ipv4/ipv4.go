package ipv4

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"github.com/point-c/wg"
	"github.com/point-c/wg/pkg/ipcheck"
	"log/slog"
	"math/rand"
	"net"
)

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
