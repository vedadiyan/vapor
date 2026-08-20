package vapor

import (
	"strings"
)

type (
	Pattern string
	Server  interface {
		Listen(addr string) error
		Shutdown() error
		HandleFunc(Pattern, func(Request) Response) error
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
