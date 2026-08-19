package vapor

import (
	"context"
	"net"
)

type (
	Pattern string
	Server  interface {
		Listen(addr net.Addr) error
		Shutdown(context.Context) error
		HandleFunc(Pattern, func(Request) (Response, error))
	}
)
