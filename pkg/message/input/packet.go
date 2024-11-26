package input

import (
	"encoding/base64"
	"github.com/beetbasket/runner/pkg/message"
	"github.com/beetbasket/runner/pkg/message/internal/kind/input"
)

type PacketInput struct {
	message.BaseMessageKind[input.Packet]
	Prefix string
	Data   message.Data `json:"data"`
}

func (pi PacketInput) Input() []byte {
	data := base64.StdEncoding.AppendEncode(append([]byte(pi.Prefix), ' '), pi.Data)
	return append(data, '\n')
}

func NewPacketInput[D message.DataLike](prefix string, data D) message.Input {
	return PacketInput{
		BaseMessageKind: message.NewBaseMessageKind[input.Packet](),
		Prefix:          prefix,
		Data:            []byte(data),
	}
}
