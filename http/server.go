package http

import (
	"context"
	"fmt"
	"io"
	"log"
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

func (srv *server) HandleFunc(pattern vapor.Pattern, fn func(vapor.Message) (vapor.Status, vapor.Message)) error {
	srv.mux.HandleFunc(string(pattern), func(w http.ResponseWriter, r *http.Request) {
		status, message := fn(newRequest(r))

		for key, values := range message.Headers() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		content, err := vapor.ReadIfNotNil(message.Content())
		if err != nil {
			log.Println(err.Error())
			w.Header().Set("X-Error", "true")
		}

		w.WriteHeader(int(status))
		_, _ = w.Write(content)
	})
	return nil
}

func (r request) Content() io.ReadCloser {
	return r.Body
}

func (r request) Context() context.Context {
	return r.Request.Context()
}

func (r request) Type() string {
	return r.Header.Get("X-Type")
}

func (r request) Subject() string {
	return r.Header.Get("X-Subject")
}

func (r request) ID() string {
	return r.Header.Get("X-ID")
}

func (r request) Headers() vapor.KeyValue {
	return *(*vapor.KeyValue)(unsafe.Pointer(&r.Header))
}

func newRequest(r *http.Request) vapor.Message {
	return &request{Request: r}
}
