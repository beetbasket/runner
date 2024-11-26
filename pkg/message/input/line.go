package input

import (
	"fmt"
	"github.com/beetbasket/runner/pkg/message"
	"github.com/beetbasket/runner/pkg/message/internal/kind/input"
)

type TextInput struct {
	message.BaseMessageKind[input.Text]
	Data message.Data `json:"data"`
}

func (ti TextInput) Input() []byte {
	return ti.Data
}

func newTextInput[D message.DataLike](data D) TextInput {
	return TextInput{Data: []byte(data)}
}

func NewInputln[D message.DataLike](data D) message.Input {
	return newTextInput(append([]byte(data), '\n'))
}

func NewInput[D message.DataLike](data D) message.Input {
	return newTextInput(data)
}

func NewInputf(format string, a ...any) message.Input {
	return newTextInput(fmt.Sprintf(format, a...))
}
