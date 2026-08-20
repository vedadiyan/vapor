package vapor

import (
	"context"
	"net"
	"net/url"
	"strings"
)

type (
	Pattern string
	Options struct {
		URI *url.URL
	}
	Option func(*Options)
	Server interface {
		Listen(addr net.Addr) error
		Shutdown(context.Context) error
		HandleFunc(Pattern, func(Request, ...Option) Response) error
	}
)

func (p Pattern) Segments() []string {
	trimmedPattern := strings.Trim(string(p), " ")
	return strings.Split(trimmedPattern, "/")
}

func (p Pattern) Tokens() map[string]int {
	out := make(map[string]int)
	for i, seg := range p.Segments() {
		if after, ok := strings.CutPrefix(seg, ":"); ok {
			out[after] = i
		}
	}

	return out
}
