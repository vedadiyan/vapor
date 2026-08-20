package http

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vedadiyan/vapor"
)

func TestServerHandleFunc(t *testing.T) {
	srv := New()

	err := srv.HandleFunc(vapor.Pattern("GET /users/:id"), func(r vapor.Request) vapor.Response {
		if r.Method() != "GET" {
			t.Errorf("method = %q", r.Method())
		}

		if r.Params()["id"] != "42" {
			t.Errorf("id = %q", r.Params()["id"])
		}

		if r.QueryString() != "foo=bar" {
			t.Errorf("query = %q", r.QueryString())
		}

		return vapor.NewResponse(
			http.StatusOK,
			vapor.WithHeaders(vapor.KeyValue{
				"Content-Type": {"text/plain"},
			}),
			vapor.WithContent([]byte("hello")),
		)
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/42?foo=bar",
		nil,
	)

	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}

	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Fatalf("content-type = %q", rec.Header().Get("Content-Type"))
	}
}

func TestServerListenAndShutdown(t *testing.T) {
	srv := New()

	if err := srv.Listen("127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	s.mut.Lock()
	if s.server == nil {
		s.mut.Unlock()
		t.Fatal("server should not be nil")
	}
	s.mut.Unlock()

	if err := srv.Shutdown(); err != nil {
		t.Fatal(err)
	}

	srv.Wait()

	s.mut.Lock()
	defer s.mut.Unlock()

	if s.server != nil {
		t.Fatal("server should be nil after shutdown")
	}
}

func TestServerAlreadyRunning(t *testing.T) {
	srv := New()

	if err := srv.Listen("127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	if err := srv.Listen("127.0.0.1:0"); err == nil {
		t.Fatal("expected already-running error")
	}
}

func TestServerOptions(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	srv := New(
		WithReadTimeout(time.Second),
		WithReadHeaderTimeout(2*time.Second),
		WithWriteTimeout(3*time.Second),
		WithIdleTimeout(4*time.Second),
		WithMaxHeaderBytes(8192),
		WithErrorLog(logger),
		WithBaseContext(func(net.Listener) context.Context {
			return context.Background()
		}),
		WithConnContext(func(ctx context.Context, _ net.Conn) context.Context {
			return ctx
		}),
		WithConnState(func(net.Conn, http.ConnState) {}),
	)

	if err := srv.Listen("127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	s := srv.(*server)

	s.mut.Lock()
	h := s.server
	s.mut.Unlock()

	if h == nil {
		t.Fatal("server should not be nil")
	}

	if h.ReadTimeout != time.Second {
		t.Errorf("ReadTimeout = %v", h.ReadTimeout)
	}

	if h.ReadHeaderTimeout != 2*time.Second {
		t.Errorf("ReadHeaderTimeout = %v", h.ReadHeaderTimeout)
	}

	if h.WriteTimeout != 3*time.Second {
		t.Errorf("WriteTimeout = %v", h.WriteTimeout)
	}

	if h.IdleTimeout != 4*time.Second {
		t.Errorf("IdleTimeout = %v", h.IdleTimeout)
	}

	if h.MaxHeaderBytes != 8192 {
		t.Errorf("MaxHeaderBytes = %d", h.MaxHeaderBytes)
	}

	if h.ErrorLog != logger {
		t.Error("ErrorLog not applied")
	}

	if h.BaseContext == nil {
		t.Error("BaseContext not applied")
	}

	if h.ConnContext == nil {
		t.Error("ConnContext not applied")
	}

	if h.ConnState == nil {
		t.Error("ConnState not applied")
	}
}

func TestRequestHeaders(t *testing.T) {
	srv := New()

	var got vapor.Request

	err := srv.HandleFunc(vapor.Pattern("GET /"), func(r vapor.Request) vapor.Response {
		got = r
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-ID", "abc")

	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if got == nil {
		t.Fatal("request was not received")
	}

	if got.ID() != "abc" {
		t.Errorf("ID = %q", got.ID())
	}

	if got.Method() != "GET" {
		t.Errorf("Method = %q", got.Method())
	}

	if got.Headers().Get("X-ID") != "abc" {
		t.Errorf("X-ID = %q", got.Headers().Get("X-ID"))
	}
}

func TestPublishOnly(t *testing.T) {
	srv := New()

	err := srv.HandleFunc(vapor.Pattern("POST /"), func(r vapor.Request) vapor.Response {
		if r.Type() != vapor.TypePublishOnly {
			t.Errorf("type = %v", r.Type())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-PublishOnly", "true")

	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)
}

func TestRequestRequiresReply(t *testing.T) {
	srv := New()

	err := srv.HandleFunc(vapor.Pattern("POST /"), func(r vapor.Request) vapor.Response {
		if r.Type() != vapor.TypeRequiresReply {
			t.Errorf("type = %v", r.Type())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)
}

func TestRequestContent(t *testing.T) {
	srv := New()

	var body string

	err := srv.HandleFunc(vapor.Pattern("POST /"), func(r vapor.Request) vapor.Response {
		data, err := io.ReadAll(r.Content())
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		body = string(data)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("hello"),
	)

	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if body != "hello" {
		t.Fatalf("body = %q", body)
	}
}

func TestRequestContext(t *testing.T) {
	srv := New()

	var got context.Context

	err := srv.HandleFunc(vapor.Pattern("GET /"), func(r vapor.Request) vapor.Response {
		got = r.Context()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if got == nil {
		t.Fatal("context should not be nil")
	}
}

func TestRequestSubject(t *testing.T) {
	srv := New()

	var subject string

	err := srv.HandleFunc(vapor.Pattern("GET /hello"), func(r vapor.Request) vapor.Response {
		subject = r.Subject()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if subject != "/hello" {
		t.Fatalf("subject = %q", subject)
	}
}

func TestRequestPattern(t *testing.T) {
	pattern := vapor.Pattern("GET /users/:id")

	srv := New()

	var got vapor.Pattern

	err := srv.HandleFunc(pattern, func(r vapor.Request) vapor.Response {
		got = r.Pattern()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/42",
		nil,
	)

	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if got != pattern {
		t.Fatalf("pattern = %q", got)
	}
}

func TestRequestQueryString(t *testing.T) {
	srv := New()

	var query vapor.QueryString

	err := srv.HandleFunc(vapor.Pattern("GET /"), func(r vapor.Request) vapor.Response {
		query = r.QueryString()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(
		http.MethodGet,
		"/?foo=bar&baz=qux",
		nil,
	)

	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if query != "foo=bar&baz=qux" {
		t.Fatalf("query = %q", query)
	}
}
