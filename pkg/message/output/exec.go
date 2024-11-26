package output

import (
	"github.com/beetbasket/runner/pkg/message"
	"github.com/beetbasket/runner/pkg/message/internal/kind/output"
)

type (
	StartMessage struct {
		message.BaseMessageKind[output.Start]
	}
	ExitMessage struct {
		message.BaseMessageKind[output.Exit]
		Code int `json:"code"`
	}
)

func NewStartMessage() message.Message {
	return StartMessage{BaseMessageKind: message.NewBaseMessageKind[output.Start]()}
}

func NewExitMessage(code int) message.Message {
	return ExitMessage{
		BaseMessageKind: message.NewBaseMessageKind[output.Exit](),
		Code:            code,
	}
}
