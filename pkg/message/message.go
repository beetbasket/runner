package message

import (
	"encoding/json"
	"fmt"
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
	JSONString[S fmt.Stringer] struct{}
	JSONKind[S fmt.Stringer]   struct {
		Kind JSONString[S] `json:"kind"`
	}
	Data []byte
)

func (JSONString[S]) MarshalJSON() ([]byte, error) {
	return json.Marshal((*new(S)).String())
}

func (d Data) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}

type DataLike interface {
	~string | ~[]byte
}

func (bm BaseMessage) Message() BaseMessage { return bm }

func NewBaseMessage() BaseMessage {
	return BaseMessage{Time: time.Now()}
}

func NewBaseMessageKind[K fmt.Stringer]() BaseMessageKind[K] {
	return BaseMessageKind[K]{BaseMessage: NewBaseMessage()}
}
