package nats

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"unsafe"

	"github.com/nats-io/nats.go"
	"github.com/vedadiyan/vapor"
)

type (
	server struct {
		server        *nats.Conn
		options       []nats.Option
		subscriptions []func(*nats.Conn) error
		mut           sync.Mutex
	}

	request struct {
		*nats.Msg
		ctx context.Context
	}
)

func (srv *server) Listen(addr net.Addr) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()
	if srv.server != nil {
		return fmt.Errorf("server is already running")
	}
	conn, err := nats.Connect(addr.String(), srv.options...)
	if err != nil {
		return err
	}
	for _, fn := range srv.subscriptions {
		if err := fn(conn); err != nil {
			conn.Close()
			return err
		}
	}
	srv.server = conn
	return nil
}

func (srv *server) Shutdown(_ context.Context) error {
	srv.mut.Lock()
	if srv.server == nil {
		return nil
	}

	conn := srv.server
	srv.server = nil
	srv.mut.Unlock()
	if err := conn.Drain(); err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (srv *server) HandleFunc(pattern vapor.Pattern, fn func(vapor.Message) (vapor.Status, vapor.Message)) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()
	subsFn := func(conn *nats.Conn) error {
		_, err := conn.Subscribe(string(pattern), func(msg *nats.Msg) {
			go func() {
				out := &nats.Msg{
					Header: make(nats.Header),
				}
				_, message := fn(newRequest(msg))

				for key, values := range message.Headers() {
					for _, value := range values {
						out.Header.Add(key, value)
					}
				}

				data, err := vapor.ReadIfNotNil(message.Content())
				if err != nil {
					log.Println(err.Error())
					out.Header.Set("X-Error", "true")
				}

				out.Data = data
				_ = msg.RespondMsg(out)
			}()
		})
		return err
	}
	srv.subscriptions = append(srv.subscriptions, subsFn)
	if srv.server != nil {
		return subsFn(srv.server)
	}
	return nil
}

func (r request) Content() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(r.Data))
}

func (r request) Context() context.Context {
	return r.ctx
}

func (r request) Type() string {
	return r.Header.Get("X-Type")
}

func (r request) Subject() string {
	return r.Msg.Subject
}

func (r request) ID() string {
	return r.Header.Get("X-ID")
}

func (r request) Headers() vapor.KeyValue {
	return *(*vapor.KeyValue)(unsafe.Pointer(&r.Header))
}

func newRequest(r *nats.Msg) vapor.Message {
	return &request{Msg: r, ctx: context.Background()}
}
