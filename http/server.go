package http

import (
	"context"
	"fmt"
	"io"
	"net"
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
	}

	request struct {
		*http.Request
	}
)

func (srv *server) Listen(addr net.Addr) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()
	if srv.server != nil {
		return fmt.Errorf("server is already running")
	}
	server := &http.Server{
		Addr:    addr.String(),
		Handler: &srv.mux,
	}
	srv.server = server
	go srv.server.ListenAndServe()
	return nil
}

func (srv *server) Shutdown(ctx context.Context) error {
	srv.mut.Lock()
	if srv.server == nil {
		return nil
	}
	ref := srv.server
	srv.server = nil
	srv.mut.Unlock()
	return ref.Shutdown(ctx)
}

func (srv *server) HandleFunc(pattern vapor.Pattern, fn func(vapor.Request) (vapor.Response, error)) error {
	srv.mux.HandleFunc(string(pattern), func(w http.ResponseWriter, r *http.Request) {
		res, err := fn(newRequest(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (r request) Content() ([]byte, error) {
	return io.ReadAll(r.Body)
}

func (r request) Context() context.Context {
	return r.Request.Context()
}

func (r request) Headers() vapor.KeyValue {
	return *(*vapor.KeyValue)(unsafe.Pointer(&r.Header))
}

func (r request) Trailers() vapor.KeyValue {
	return *(*vapor.KeyValue)(unsafe.Pointer(&r.Trailer))
}

func newRequest(r *http.Request) vapor.Request {
	return &request{Request: r}
}
