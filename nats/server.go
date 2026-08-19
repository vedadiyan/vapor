package nats

import (
	"context"
	"net"
	"strconv"
	"unsafe"

	"github.com/nats-io/nats.go"
	"github.com/vedadiyan/vapor"
)

type (
	server struct {
		server        *nats.Conn
		options       []nats.Option
		subscriptions chan func(*nats.Conn)
	}

	request struct {
		*nats.Msg
		ctx context.Context
	}
)

func (srv *server) Listen(addr net.Addr) error {
	conn, err := nats.Connect(addr.String(), srv.options...)
	if err != nil {
		return err
	}
	srv.server = conn
	go func() {
		for fn := range srv.subscriptions {
			fn(conn)
		}
	}()
	return nil
}

func (srv *server) Shutdown(ctx context.Context) error {
	if srv.server == nil {
		return nil
	}
	close(srv.subscriptions)

	done := make(chan error, 1)

	go func() {
		done <- srv.server.Drain()
	}()

	select {
	case err := <-done:
		srv.server.Close()
		return err
	case <-ctx.Done():
		srv.server.Close()
		return ctx.Err()
	}
}

func (srv *server) HandleFunc(pattern vapor.Pattern, fn func(vapor.Request) (vapor.Response, error)) error {
	srv.subscriptions <- func(conn *nats.Conn) {
		_, _ = conn.Subscribe(string(pattern), func(msg *nats.Msg) {
			go func() {
				out := &nats.Msg{
					Header: make(nats.Header),
				}
				res, err := fn(newRequest(msg))
				if err != nil {
					out.Header.Set("Status", "500")
					out.Data = []byte(err.Error())
					_ = msg.RespondMsg(out)
					return
				}
				out.Header.Set("Status", strconv.Itoa(res.Status()))

				for key, values := range res.Headers() {
					for _, value := range values {
						out.Header.Add(key, value)
					}
				}
				out.Data = res.Content()
				_ = msg.RespondMsg(out)
			}()
		})
	}
	return nil
}

func (r request) Content() ([]byte, error) {
	return r.Msg.Data, nil
}

func (r request) Context() context.Context {
	return r.ctx
}

func (r request) Headers() vapor.KeyValue {
	return *(*vapor.KeyValue)(unsafe.Pointer(&r.Header))
}

func (r request) Trailers() vapor.KeyValue {
	return nil
}

func newRequest(r *nats.Msg) vapor.Request {
	return &request{Msg: r, ctx: context.Background()}
}
