package nats

import (
	"io"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/vedadiyan/vapor"
)

func BenchmarkHandleFunc(b *testing.B) {
	srv := New()

	if err := srv.HandleFunc(vapor.Pattern("users/:id"), func(r vapor.Request) vapor.Response {
		return vapor.NewResponse(
			200,
			vapor.WithContent([]byte("hello")),
		)
	}); err != nil {
		b.Fatal(err)
	}

	msg := &nats.Msg{
		Subject: "users.42",
		Header:  nats.Header{},
		Data:    []byte("hello"),
	}

	msg.Header.Set("X-Method", "GET")

	pattern := vapor.Pattern("users/:id")
	tokens := pattern.Tokens()

	b.ReportAllocs()

	for b.Loop() {
		req := newRequest(msg, pattern, tokens)
		_ = req.Subject()
	}
}

func BenchmarkRequestParams(b *testing.B) {
	pattern := vapor.Pattern("users/:id")

	msg := &nats.Msg{
		Subject: "users.42",
	}

	req := newRequest(
		msg,
		pattern,
		pattern.Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Params()
	}
}

func BenchmarkRequestHeaders(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
		Header:  nats.Header{},
	}

	msg.Header.Set("X-ID", "123")
	msg.Header.Set("X-Method", "GET")
	msg.Header.Set("Authorization", "Bearer token")
	msg.Header.Set("Content-Type", "application/json")

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Headers()
	}
}

func BenchmarkRequestContent(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
		Data:    []byte("hello"),
	}

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		body := req.Content()
		_, _ = io.ReadAll(body)
		_ = body.Close()
	}
}

func BenchmarkRequestMethod(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
		Header:  nats.Header{},
	}

	msg.Header.Set("X-Method", "GET")

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Method()
	}
}

func BenchmarkRequestID(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
		Header:  nats.Header{},
	}

	msg.Header.Set("X-ID", "abc")

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.ID()
	}
}

func BenchmarkRequestQueryString(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
		Header:  nats.Header{},
	}

	msg.Header.Set("X-Q", "foo=bar&baz=qux")

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.QueryString()
	}
}

func BenchmarkRequestSubject(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
	}

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Subject()
	}
}

func BenchmarkRequestPattern(b *testing.B) {
	pattern := vapor.Pattern("users/:id")

	msg := &nats.Msg{
		Subject: "users.42",
	}

	req := newRequest(
		msg,
		pattern,
		pattern.Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Pattern()
	}
}

func BenchmarkRequestContext(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
	}

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Context()
	}
}

func BenchmarkRequestType(b *testing.B) {
	msg := &nats.Msg{
		Subject: "users.42",
		Reply:   "_INBOX.reply",
	}

	req := newRequest(
		msg,
		vapor.Pattern("users/:id"),
		vapor.Pattern("users/:id").Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Type()
	}
}

func BenchmarkFullRequest(b *testing.B) {
	pattern := vapor.Pattern("users/:id")

	msg := &nats.Msg{
		Subject: "users.42",
		Header:  nats.Header{},
		Data:    []byte("hello"),
	}

	msg.Header.Set("X-ID", "abc")
	msg.Header.Set("X-Method", "GET")
	msg.Header.Set("X-Q", "foo=bar")

	req := newRequest(
		msg,
		pattern,
		pattern.Tokens(),
	)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Method()
		_ = req.ID()
		_ = req.Params()
		_ = req.QueryString()
		_ = req.Headers()
		_ = req.Subject()
		_ = req.Pattern()
		_ = req.Context()
		_ = req.Type()

		body := req.Content()
		_, _ = io.ReadAll(body)
		_ = body.Close()
	}
}

func BenchmarkHandleFuncRequest(b *testing.B) {
	pattern := vapor.Pattern("users/:id")
	tokens := pattern.Tokens()

	msg := &nats.Msg{
		Subject: "users.42",
		Header:  nats.Header{},
		Data:    []byte("hello"),
	}

	msg.Header.Set("X-ID", "abc")
	msg.Header.Set("X-Method", "GET")
	msg.Header.Set("X-Q", "foo=bar")

	handler := func(r vapor.Request) vapor.Response {
		_ = r.Method()
		_ = r.ID()
		_ = r.Params()
		_ = r.QueryString()
		_ = r.Headers()
		_ = r.Subject()

		return vapor.NewResponse(
			200,
			vapor.WithContent([]byte("hello")),
		)
	}

	b.ReportAllocs()

	for b.Loop() {
		req := newRequest(msg, pattern, tokens)
		_ = handler(req)
	}
}
