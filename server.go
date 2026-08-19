package vapor

import (
	"context"
	"net"
)

type (
	Status  int
	Pattern string
	Server  interface {
		Listen(addr net.Addr) error
		Shutdown(context.Context) error
		HandleFunc(Pattern, func(Message) (Status, Message)) error
	}
)
