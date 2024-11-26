package stdio

import (
	"github.com/beetbasket/runner/pkg/message/internal/kind"
)

type (
	Stderr = kind.Kind[stderr]
	Stdout = kind.Kind[stdout]
	Stdin  = kind.Kind[stdin]
)

type (
	stderr struct{}
	stdout struct{}
	stdin  struct{}
)
