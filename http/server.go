package http

import (
	"context"
	"io"
	"net"
	"net/http"
	"unsafe"

	"github.com/vedadiyan/vapor"
)

type (
	server struct {
		mux    http.ServeMux
		server http.Server
	}

	request struct {
		*http.Request
	}
)

func (srv *server) Listen(addr net.Addr) error {
	ln, err := net.Listen("tcp", addr.String())
	if err != nil {
		return err
	}

	return srv.server.Serve(ln)
}

func (srv *server) Shutdown(ctx context.Context) error {
	return srv.server.Shutdown(ctx)
}

func (srv *server) HandleFunc(pattern vapor.Pattern, fn func(vapor.Request) (vapor.Response, error)) {
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
