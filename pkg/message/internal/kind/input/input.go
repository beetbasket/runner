package input

import (
	"github.com/beetbasket/runner/pkg/message/internal/kind"
)

type (
	Text   = kind.Kind[text]
	Packet = kind.Kind[packet]
)

type (
	text   struct{}
	packet struct{}
)
