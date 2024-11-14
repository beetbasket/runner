package runner

import (
	"fmt"
	"github.com/beetbasket/runner/internal/kind/input"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

var editorPrefix = "()()()!@#$%%^&*(32u4io32u4i"

type Input interface {
	Input() []byte
}

type LineInput struct {
	BaseMessageKind[input.Line]
	Data []byte `json:"data"`
}

func (li LineInput) Input() []byte {
	return append(slices.Clone(li.Data), '\n')
}

type EditInput struct {
	BaseMessageKind[input.Editor]
	ID   uuid.UUID `json:"id"`
	Data []byte    `json:"data"`
}

func (ei EditInput) Input() []byte {
	return []byte(fmt.Sprintf("%s %s %s\n", editorPrefix, ei.ID, ei.Data))
}

func NewLineInput[D dataLike](data D) LineInput {
	return LineInput{
		BaseMessageKind: NewBaseMessageKind[input.Line](),
		Data:            []byte(data),
	}
}

func NewEditInput[D dataLike](id uuid.UUID, data D) EditInput {
	return EditInput{
		BaseMessageKind: NewBaseMessageKind[input.Editor](),
		ID:              id,
		Data:            []byte(data),
	}
}
