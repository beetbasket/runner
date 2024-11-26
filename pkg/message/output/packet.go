package output

import (
	"github.com/beetbasket/runner/pkg/message"
	"github.com/beetbasket/runner/pkg/message/internal/kind/output"
)

type PacketMessage struct {
	message.BaseMessageKind[output.Packet]
	Data message.Data `json:"data"`
}

func NewPacketMessage[D message.DataLike](data D) message.Message {
	return PacketMessage{
		BaseMessageKind: message.NewBaseMessageKind[output.Packet](),
		Data:            message.Data(data),
	}
}
