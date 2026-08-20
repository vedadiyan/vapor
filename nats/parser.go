package nats

import (
	"bytes"
	"strings"

	"github.com/vedadiyan/vapor"
)

func toSubject(pattern vapor.Pattern) string {
	builder := bytes.NewBufferString("")

	for i, seg := range pattern.Segments() {
		if i > 0 {
			builder.WriteRune('.')
		}
		if strings.HasPrefix(seg, ":") {
			builder.WriteRune('*')
			continue
		}
		builder.WriteString(seg)
	}

	return builder.String()
}

func getParams(subject string, tokens map[string]int) map[string]string {
	subjectTokens := strings.Split(subject, ".")
	l := len(subjectTokens)
	out := make(map[string]string)
	for key, value := range tokens {
		if value >= l {
			continue
		}
		out[key] = subjectTokens[value]
	}
	return out
}
