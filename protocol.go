package vapor

import (
	"context"
	"io"
)

type (
	KeyValue map[string][]string
	Message  interface {
		Content() io.ReadCloser
		Context() context.Context
		Subject() string
		Type() string
		ID() string
		Headers() KeyValue
	}
)
