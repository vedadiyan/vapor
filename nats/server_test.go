package nats

import (
	"context"
	"io"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/vedadiyan/vapor"
)

func startTestServer(t *testing.T) *natsserver.Server {
	t.Helper()

	opts := &natsserver.Options{
		Host:   "127.0.0.1",
		Port:   -1,
		NoLog:  true,
		NoSigs: true,
	}

	ns, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}

	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		ns.Shutdown()
		t.Fatal("NATS server did not become ready")
	}

	t.Cleanup(ns.Shutdown)

	return ns
}

func TestServerHandleFunc(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	err := srv.HandleFunc(vapor.Pattern("users/:id"), func(r vapor.Request) vapor.Response {
		if r.Subject() != "users.42" {
			t.Errorf("subject = %q", r.Subject())
		}

		if r.Params()["id"] != "42" {
			t.Errorf("id = %q", r.Params()["id"])
		}

		if r.Method() != "GET" {
			t.Errorf("method = %q", r.Method())
		}

		if r.QueryString() != "foo=bar" {
			t.Errorf("query = %q", r.QueryString())
		}

		return vapor.NewResponse(
			200,
			vapor.WithHeaders(vapor.KeyValue{
				"Content-Type": {"text/plain"},
			}),
			vapor.WithContent([]byte("hello")),
		)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg := &nats.Msg{
		Subject: "users.42",
		Data:    []byte("request"),
		Header:  nats.Header{},
	}

	msg.Header.Set("X-Method", "GET")
	msg.Header.Set("X-Q", "foo=bar")

	reply, err := client.RequestMsg(msg, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if string(reply.Data) != "hello" {
		t.Fatalf("body = %q", reply.Data)
	}

	if reply.Header.Get("Content-Type") != "text/plain" {
		t.Fatalf("content-type = %q", reply.Header.Get("Content-Type"))
	}

	if reply.Header.Get("X-Status") != "200" {
		t.Fatalf("status = %q", reply.Header.Get("X-Status"))
	}
}

func TestServerListenAndShutdown(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}

	srv.mut.Lock()
	if srv.server == nil {
		srv.mut.Unlock()
		t.Fatal("server should not be nil")
	}
	srv.mut.Unlock()

	if err := srv.Shutdown(); err != nil {
		t.Fatal(err)
	}

	srv.Wait()

	srv.mut.Lock()
	defer srv.mut.Unlock()

	if srv.server != nil {
		t.Fatal("server should be nil after shutdown")
	}
}

func TestServerAlreadyRunning(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	if err := srv.Listen(ns.ClientURL()); err == nil {
		t.Fatal("expected already-running error")
	}
}

func TestServerHandleFuncAfterListen(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	if err := srv.HandleFunc(vapor.Pattern("hello"), func(r vapor.Request) vapor.Response {
		return vapor.NewResponse(
			200,
			vapor.WithContent([]byte("hello")),
		)
	}); err != nil {
		t.Fatal(err)
	}

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg, err := client.Request("hello", nil, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if string(msg.Data) != "hello" {
		t.Fatalf("body = %q", msg.Data)
	}
}

func TestRequestHeaders(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	var got vapor.Request
	done := make(chan struct{})

	err := srv.HandleFunc(vapor.Pattern("test"), func(r vapor.Request) vapor.Response {
		got = r
		close(done)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg := &nats.Msg{
		Subject: "test",
		Data:    []byte("hello"),
		Header:  nats.Header{},
	}

	msg.Header.Set("X-ID", "abc")
	msg.Header.Set("X-Method", "GET")
	msg.Header.Set("X-Q", "foo=bar")

	if err := client.PublishMsg(msg); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}

	if got == nil {
		t.Fatal("request was not received")
	}

	if got.ID() != "abc" {
		t.Errorf("ID = %q", got.ID())
	}

	if got.Method() != "GET" {
		t.Errorf("Method = %q", got.Method())
	}

	if got.QueryString() != "foo=bar" {
		t.Errorf("QueryString = %q", got.QueryString())
	}

	if got.Headers().Get("X-ID") != "abc" {
		t.Errorf("X-ID = %q", got.Headers().Get("X-ID"))
	}
}

func TestPublishOnly(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	done := make(chan struct{})

	err := srv.HandleFunc(vapor.Pattern("events"), func(r vapor.Request) vapor.Response {
		if r.Type() != vapor.TypePublishOnly {
			t.Errorf("type = %v", r.Type())
		}

		close(done)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Publish("events", []byte("hello")); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}
}

func TestRequestRequiresReply(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	err := srv.HandleFunc(vapor.Pattern("request"), func(r vapor.Request) vapor.Response {
		if r.Type() != vapor.TypeRequiresReply {
			t.Errorf("type = %v", r.Type())
		}

		return vapor.NewResponse(
			200,
			vapor.WithContent([]byte("ok")),
		)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg, err := client.Request("request", nil, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if string(msg.Data) != "ok" {
		t.Fatalf("body = %q", msg.Data)
	}
}

func TestRequestContent(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	var body string
	done := make(chan struct{})

	err := srv.HandleFunc(vapor.Pattern("echo"), func(r vapor.Request) vapor.Response {
		data, err := io.ReadAll(r.Content())
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		body = string(data)
		close(done)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Publish("echo", []byte("hello")); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}

	if body != "hello" {
		t.Fatalf("body = %q", body)
	}
}

func TestRequestContext(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	var got context.Context
	done := make(chan struct{})

	err := srv.HandleFunc(vapor.Pattern("context"), func(r vapor.Request) vapor.Response {
		got = r.Context()
		close(done)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Publish("context", nil); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}

	if got == nil {
		t.Fatal("context should not be nil")
	}
}

func TestRequestSubject(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	var subject string
	done := make(chan struct{})

	err := srv.HandleFunc(vapor.Pattern("hello"), func(r vapor.Request) vapor.Response {
		subject = r.Subject()
		close(done)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Publish("hello", nil); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}

	if subject != "hello" {
		t.Fatalf("subject = %q", subject)
	}
}

func TestRequestPattern(t *testing.T) {
	ns := startTestServer(t)

	pattern := vapor.Pattern("users/:id")

	srv := New().(*server)

	var got vapor.Pattern
	done := make(chan struct{})

	err := srv.HandleFunc(pattern, func(r vapor.Request) vapor.Response {
		got = r.Pattern()
		close(done)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Publish("users.42", nil); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}

	if got != pattern {
		t.Fatalf("pattern = %q", got)
	}
}

func TestRequestQueryString(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	var query vapor.QueryString
	done := make(chan struct{})

	err := srv.HandleFunc(vapor.Pattern("query"), func(r vapor.Request) vapor.Response {
		query = r.QueryString()
		close(done)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg := &nats.Msg{
		Subject: "query",
		Header:  nats.Header{},
	}

	msg.Header.Set("X-Q", "foo=bar&baz=qux")

	if err := client.PublishMsg(msg); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}

	if query != "foo=bar&baz=qux" {
		t.Fatalf("query = %q", query)
	}
}

func TestRequestID(t *testing.T) {
	ns := startTestServer(t)

	srv := New().(*server)

	var id string
	done := make(chan struct{})

	err := srv.HandleFunc(vapor.Pattern("request"), func(r vapor.Request) vapor.Response {
		id = r.ID()
		close(done)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Listen(ns.ClientURL()); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	client, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg := &nats.Msg{
		Subject: "request",
		Header:  nats.Header{},
	}

	msg.Header.Set("X-ID", "abc")

	if err := client.PublishMsg(msg); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler")
	}

	if id != "abc" {
		t.Fatalf("ID = %q", id)
	}
}
