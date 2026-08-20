package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"unsafe"

	"github.com/vedadiyan/vapor"
)

type (
	Option func(*http.Server)
	server struct {
		mux     http.ServeMux
		server  *http.Server
		mut     sync.Mutex
		wg      sync.WaitGroup
		options []Option
	}

	request struct {
		*http.Request
		pattern vapor.Pattern
		tokens  map[string]int
	}
)

func New(opts ...Option) vapor.Server {
	return &server{options: opts}
}

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

	for _, opt := range srv.options {
		opt(server)
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}

	if server.TLSConfig != nil {
		ln = tls.NewListener(ln, server.TLSConfig)
	}

	srv.server = server
	srv.wg.Add(1)

	go func() {
		defer srv.wg.Done()
		go func() {
			defer srv.wg.Done()

			if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
				if server.ErrorLog != nil {
					server.ErrorLog.Printf("server error: %v", err)
				}
			}
		}()
	}()

	return nil
}

func (srv *server) Shutdown() error {
	srv.mut.Lock()
	if srv.server == nil {
		return nil
	}
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
			w.WriteHeader(http.StatusAccepted)
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
	if r.Request.Header.Get("X-PublishOnly") == "true" {
		return vapor.TypePublishOnly
	}
	return vapor.TypeRequiresReply
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

func WithReadTimeout(timeout time.Duration) Option {
	return func(s *http.Server) {
		s.ReadTimeout = timeout
	}
}

func WithReadHeaderTimeout(timeout time.Duration) Option {
	return func(s *http.Server) {
		s.ReadHeaderTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *http.Server) {
		s.WriteTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) Option {
	return func(s *http.Server) {
		s.IdleTimeout = timeout
	}
}

func WithMaxHeaderBytes(n int) Option {
	return func(s *http.Server) {
		s.MaxHeaderBytes = n
	}
}

func WithErrorLog(logger *log.Logger) Option {
	return func(s *http.Server) {
		s.ErrorLog = logger
	}
}

func WithTLSConfig(config *tls.Config) Option {
	return func(s *http.Server) {
		s.TLSConfig = config
	}
}

func WithBaseContext(fn func(net.Listener) context.Context) Option {
	return func(s *http.Server) {
		s.BaseContext = fn
	}
}

func WithConnContext(fn func(ctx context.Context, c net.Conn) context.Context) Option {
	return func(s *http.Server) {
		s.ConnContext = fn
	}
}

func WithConnState(fn func(net.Conn, http.ConnState)) Option {
	return func(s *http.Server) {
		s.ConnState = fn
	}
}
