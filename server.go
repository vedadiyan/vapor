package vapor

import (
	"context"
	"net"
	"net/url"
)

type (
	Status  int
	Pattern string
	Options struct {
		URI *url.URL
	}
	Option func(*Options)
	Server interface {
		Listen(addr net.Addr) error
		Shutdown(context.Context) error
		HandleFunc(Pattern, func(Message, ...Option) (Status, Message)) error
	}
)
