package vapor

import (
	"context"
	"io"
)

type (
	KeyValue    map[string][]string
	ParamStore  map[string]string
	QueryString string
	Message     interface {
		Content() io.ReadCloser
		Context() context.Context
		Subject() string
		Type() string
		ID() string
		Method() string
		Params() ParamStore
		QueryString() QueryString
		Headers() KeyValue
	}
)
