package nats

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
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
		wg            sync.WaitGroup
	}

	request struct {
		*nats.Msg
		tokens  map[string]int
		pattern vapor.Pattern
		ctx     context.Context
	}
)

func New(opts ...nats.Option) vapor.Server {
	return &server{options: opts}
}

func (srv *server) Listen(addr string) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()
	if srv.server != nil {
		return fmt.Errorf("server is already running")
	}
	conn, err := nats.Connect(addr, srv.options...)
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
	conn := srv.server
	srv.server = nil
	srv.mut.Unlock()
	if err := conn.Drain(); err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (srv *server) Wait() {
	srv.wg.Wait()
}

func (srv *server) HandleFunc(pattern vapor.Pattern, fn func(vapor.Request) vapor.Response) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	tokens := pattern.Tokens()
	subsFn := func(conn *nats.Conn) error {
		_, err := conn.Subscribe(toSubject(pattern), func(msg *nats.Msg) {
			go func() {
				out := &nats.Msg{
					Header: make(nats.Header),
				}

				res := fn(newRequest(msg, pattern, tokens))
				if res == nil {
					return
				}

				for key, values := range res.Headers() {
					for _, value := range values {
						out.Header.Add(key, value)
					}
				}

				out.Header.Add("X-Status", strconv.Itoa(res.Status()))

				out.Data = res.Content()
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

func (r request) Type() vapor.Type {
	if len(r.Msg.Reply) != 0 {
		return vapor.TypeRequiresReply
	}
	return vapor.TypePublishOnly
}

func (r request) Subject() string {
	return r.Msg.Subject
}

func (r request) Method() string {
	return r.Msg.Header.Get("X-Method")
}

func (r request) Params() vapor.ParamStore {
	return getParams(r.Msg.Subject, r.tokens)
}

func (r request) Pattern() vapor.Pattern {
	return r.pattern
}

func (r request) QueryString() vapor.QueryString {
	return vapor.QueryString(r.Msg.Header.Get("X-Q"))
}

func (r request) ID() string {
	return r.Header.Get("X-ID")
}

func (r request) Headers() vapor.KeyValue {
	return *(*vapor.KeyValue)(unsafe.Pointer(&r.Header))
}

func newRequest(r *nats.Msg, patten vapor.Pattern, tokens map[string]int) vapor.Request {
	return &request{Msg: r, tokens: tokens, pattern: patten, ctx: context.Background()}
}
