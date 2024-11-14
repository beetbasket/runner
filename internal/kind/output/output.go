package output

import "github.com/beetbasket/runner/internal/kind"

type (
	Edit  = kind.Kind[edit]
	Stdio = kind.Kind[stdio]
	Start = kind.Kind[start]
	Exit  = kind.Kind[exit]
)

type (
	edit  struct{}
	stdio struct{}
	start struct{}
	exit  struct{}
)
