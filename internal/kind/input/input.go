package input

import "github.com/beetbasket/runner/internal/kind"

type (
	Line   = kind.Kind[line]
	Editor = kind.Kind[editor]
)

type (
	line   struct{}
	editor struct{}
)
