package vapor

import "context"

type (
	KeyValue map[string][]string
	Request  interface {
		Content() ([]byte, error)
		Context() context.Context
		Headers() KeyValue
		Trailers() KeyValue
	}
	Response interface {
		Status() int
		Content() []byte
		Headers() KeyValue
	}
)
