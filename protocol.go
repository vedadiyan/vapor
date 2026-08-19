package vapor

import "context"

type (
	KeyValue map[string][]string
	Message  interface {
		Content() ([]byte, error)
		Context() context.Context
		Subject() string
		Type() string
		ID() string
		Headers() KeyValue
	}
)
