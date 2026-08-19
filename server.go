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
		HandleFunc(Pattern, func(Message) (Message, error)) error
	}
)
