package edit

import "github.com/beetbasket/runner/internal/kind"

type (
	Start = kind.Kind[start]
	Stop  = kind.Kind[stop]
	Reply = kind.Kind[reply]
)

type (
	start struct{}
	reply struct{}
	stop  struct{}
)
