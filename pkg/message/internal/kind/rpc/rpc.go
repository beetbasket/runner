package rpc

import (
	"github.com/beetbasket/runner/pkg/message/internal/kind"
)

type (
	Request  = kind.Kind[response]
	Response = kind.Kind[reply]
)

type (
	response struct{}
	reply    struct{}
)
