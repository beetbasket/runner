package output

import (
	"fmt"
	"github.com/beetbasket/runner/pkg/message"
	"github.com/beetbasket/runner/pkg/message/internal/kind/output"
	"github.com/beetbasket/runner/pkg/message/internal/kind/stdio"
)

type (
	StdioMessage[K fmt.Stringer] struct {
		message.BaseMessageKind[output.Stdio]
		Stdio message.JSONString[K] `json:"stdio"`
		Data  message.Data          `json:"data"`
	}
	StdinMessage struct {
		StdioMessage[stdio.Stdin]
	}
	StderrMessage struct {
		StdioMessage[stdio.Stderr]
	}
	StdoutMessage struct {
		StdioMessage[stdio.Stdout]
	}
)

func newStdioMessage[K fmt.Stringer, D message.DataLike](data D) StdioMessage[K] {
	return StdioMessage[K]{
		BaseMessageKind: message.NewBaseMessageKind[output.Stdio](),
		Data:            message.Data(data),
	}
}

type StdioLike interface {
	StderrMessage | StdoutMessage | StdinMessage
}

func NewStdioMessage[T StdioLike, D message.DataLike](data D) message.Message {
	var msg T
	switch msg := any(&msg).(type) {
	case *StderrMessage:
		msg.StdioMessage = newStdioMessage[stdio.Stderr](data)
	case *StdoutMessage:
		msg.StdioMessage = newStdioMessage[stdio.Stdout](data)
	case *StdinMessage:
		msg.StdioMessage = newStdioMessage[stdio.Stdin](data)
	default:
		panic("invalid stdio type")
	}
	return any(msg).(message.Message)
}
