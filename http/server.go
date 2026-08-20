package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"unsafe"

	"github.com/vedadiyan/vapor"
)

type (
	server struct {
		mux    http.ServeMux
		server *http.Server
		mut    sync.Mutex
		wg     sync.WaitGroup
	}

	request struct {
		*http.Request
		pattern vapor.Pattern
		tokens  map[string]int
	}
)

func (srv *server) Listen(addr string) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()
	if srv.server != nil {
		return fmt.Errorf("server is already running")
	}
	server := &http.Server{
		Addr:    addr,
		Handler: &srv.mux,
	}
	srv.server = server
	go server.ListenAndServe()
	srv.wg.Add(1)
	return nil
}

func (srv *server) Shutdown() error {
	srv.mut.Lock()
	if srv.server == nil {
		return nil
	}
	defer func() {
		srv.wg.Done()
	}()
	ref := srv.server
	srv.server = nil
	srv.mut.Unlock()
	return ref.Shutdown(context.Background())
}

func (srv *server) Wait() {
	srv.wg.Wait()
}

func (srv *server) HandleFunc(pattern vapor.Pattern, fn func(vapor.Request) vapor.Response) error {
	tokens := pattern.Tokens()
	srv.mux.HandleFunc(string(pattern), func(w http.ResponseWriter, r *http.Request) {
		res := fn(newRequest(r, pattern, tokens))
		if res == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for key, values := range res.Headers() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(res.Status())
		_, _ = w.Write(res.Content())
	})
	return nil
}

func (r request) Content() io.ReadCloser {
	return r.Body
}

func (r request) Context() context.Context {
	return r.Request.Context()
}

func (r request) Type() vapor.Type {
	return 0
}

func (r request) Subject() string {
	return r.RequestURI
}

func (r request) Method() string {
	return r.Request.Method
}
func (r request) ID() string {
	return r.Header.Get("X-ID")
}

func (r request) Params() vapor.ParamStore {
	out := make(vapor.ParamStore)
	for key := range r.tokens {
		out[key] = r.Request.PathValue(key)
	}
	return out
}

func (r request) Pattern() vapor.Pattern {
	return r.pattern
}

func (r request) QueryString() vapor.QueryString {
	return vapor.QueryString(r.Request.URL.RawQuery)
}

func (r request) Headers() vapor.KeyValue {
	return *(*vapor.KeyValue)(unsafe.Pointer(&r.Header))
}

func newRequest(r *http.Request, pattern vapor.Pattern, tokens map[string]int) vapor.Request {
	return &request{Request: r, pattern: pattern, tokens: tokens}
}
