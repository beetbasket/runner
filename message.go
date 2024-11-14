package runner

import (
	"fmt"
	"github.com/beetbasket/runner/internal/kind/edit"
	"github.com/beetbasket/runner/internal/kind/output"
	"github.com/beetbasket/runner/internal/kind/stdio"
	"github.com/google/uuid"
	"time"
)

type Message interface {
	Message() BaseMessage
}

type BaseMessage struct {
	Time time.Time `json:"time"`
}

type BaseMessageKind[K fmt.Stringer] struct {
	BaseMessage
	JSONKind[K]
}

type (
	StartMessage struct {
		BaseMessageKind[output.Start]
	}
	ExitMessage struct {
		BaseMessageKind[output.Exit]
		Code int `json:"code"`
	}
)

type (
	StdioMessage[K fmt.Stringer] struct {
		BaseMessageKind[output.Stdio]
		Stdio JSONString[K] `json:"stdio"`
		Data  []byte        `json:"data"`
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

type (
	EditMessage[K fmt.Stringer] struct {
		BaseMessageKind[output.Edit]
		Edit   JSONString[K] `json:"edit"`
		EditID uuid.UUID     `json:"editID"`
	}
	EditStartMessage struct {
		EditMessage[edit.Start]
		Data     []byte `json:"data"`
		Filename string `json:"filename"`
	}
	EditReplyMessage struct {
		EditMessage[edit.Reply]
		Data []byte `json:"data"`
	}
	EditStopMessage struct {
		EditMessage[edit.Stop]
		Code int `json:"code"`
	}
)

type (
	JSONString[S fmt.Stringer] struct{}
	JSONKind[S fmt.Stringer]   struct {
		Kind JSONString[S] `json:"kind"`
	}
)

type dataLike interface {
	~string | ~[]byte
}

func (bm BaseMessage) Message() BaseMessage { return bm }

func NewBaseMessage() BaseMessage {
	return BaseMessage{Time: time.Now()}
}

func NewBaseMessageKind[K fmt.Stringer]() BaseMessageKind[K] {
	return BaseMessageKind[K]{BaseMessage: NewBaseMessage()}
}

func newStdioMessage[K fmt.Stringer, D dataLike](data D) StdioMessage[K] {
	return StdioMessage[K]{
		BaseMessageKind: NewBaseMessageKind[output.Stdio](),
		Data:            []byte(data),
	}
}

func NewStdioMessage[T StderrMessage | StdoutMessage | StdinMessage, D dataLike](data D) (msg T) {
	switch msg := any(&msg).(type) {
	case *StderrMessage:
		msg.StdioMessage = newStdioMessage[stdio.Stderr](data)
	case *StdoutMessage:
		msg.StdioMessage = newStdioMessage[stdio.Stdout](data)
	case *StdinMessage:
		msg.StdioMessage = newStdioMessage[stdio.Stdin](data)
	default:
		panic("here be dragons")
	}
	return
}

func NewStartMessage() StartMessage {
	return StartMessage{BaseMessageKind: NewBaseMessageKind[output.Start]()}
}

func NewStopMessage(code int) ExitMessage {
	return ExitMessage{
		BaseMessageKind: NewBaseMessageKind[output.Exit](),
		Code:            code,
	}
}

func newEditMessage[K fmt.Stringer](id uuid.UUID) EditMessage[K] {
	return EditMessage[K]{
		BaseMessageKind: NewBaseMessageKind[output.Edit](),
		EditID:          id,
	}
}

func NewEditStartMessage[D dataLike](id uuid.UUID, filename string, data D) EditStartMessage {
	return EditStartMessage{
		EditMessage: newEditMessage[edit.Start](id),
		Data:        []byte(data),
		Filename:    "filename",
	}
}

func NewEditReplyMessage[D dataLike](id uuid.UUID, data D) EditReplyMessage {
	return EditReplyMessage{
		EditMessage: newEditMessage[edit.Reply](id),
		Data:        []byte(data),
	}
}

func NewEditStopMessage(id uuid.UUID, code int) EditStopMessage {
	return EditStopMessage{
		EditMessage: newEditMessage[edit.Stop](id),
		Code:        code,
	}
}
