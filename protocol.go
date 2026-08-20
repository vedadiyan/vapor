package vapor

import (
	"context"
	"io"
)

type (
	Type        int
	KeyValue    map[string][]string
	ParamStore  map[string]string
	QueryString string
	Request     interface {
		Content() io.ReadCloser
		Context() context.Context
		Subject() string
		Type() Type
		ID() string
		Method() string
		Params() ParamStore
		Pattern() Pattern
		QueryString() QueryString
		Headers() KeyValue
	}
	Response interface {
		Content() []byte
		Context() context.Context
		Status() int
		Headers() KeyValue
	}
)

const (
	TypePublishOnly Type = iota
	TypeRequiresReply
)
