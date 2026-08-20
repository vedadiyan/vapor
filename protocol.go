package vapor

import (
	"context"
	"io"
	"strings"
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

	response struct {
		content []byte
		ctx     context.Context
		status  int
		headers KeyValue
	}

	ResponseOption func(*response)
)

const (
	TypePublishOnly Type = iota
	TypeRequiresReply
)

func (kv KeyValue) Add(key string, value string) {
	k := strings.ToLower(key)
	if _, ok := kv[k]; !ok {
		kv[k] = make([]string, 0)
	}
	kv[k] = append(kv[k], value)
}

func (kv KeyValue) Set(key string, value string) {
	k := strings.ToLower(key)
	kv[k] = append([]string{}, value)
}

func (kv KeyValue) Remove(key string) {
	k := strings.ToLower(key)
	delete(kv, k)
}

func (kv KeyValue) Get(key string) string {
	k := strings.ToLower(key)
	return strings.Join(kv[k], ",")
}

func NewResponse(status int, opts ...ResponseOption) Response {
	out := &response{status: status, ctx: context.Background()}
	for _, opt := range opts {
		opt(out)
	}

	return out
}

func (r response) Content() []byte {
	return r.content
}
func (r response) Context() context.Context {
	return r.ctx
}
func (r response) Status() int {
	return r.status
}
func (r response) Headers() KeyValue {
	return r.headers
}

func WithContent(content []byte) ResponseOption {
	return func(r *response) {
		r.content = content
	}
}

func WithHeaders(headers KeyValue) ResponseOption {
	return func(r *response) {
		r.headers = headers
	}
}

func WithContext(ctx context.Context) ResponseOption {
	return func(r *response) {
		r.ctx = ctx
	}
}
