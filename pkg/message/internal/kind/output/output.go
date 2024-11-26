package output

import (
	"github.com/beetbasket/runner/pkg/message/internal/kind"
)

type (
	Packet = kind.Kind[packet]
	Stdio  = kind.Kind[stdio]
	Start  = kind.Kind[start]
	Exit   = kind.Kind[exit]
)

type (
	packet struct{}
	stdio  struct{}
	start  struct{}
	exit   struct{}
)
