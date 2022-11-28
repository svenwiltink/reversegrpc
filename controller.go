package reversegrpc

import (
	"context"
	"errors"
	"net"
	"sync"

	"google.golang.org/grpc"
)

var ErrWorkerDisconnected = errors.New("worker disconnected")

type Controller struct {
	l    net.Listener
	addr string
}

func (c *Controller) Listen(address string) error {
	var err error
	c.l, err = net.Listen("tcp", address)
	c.addr = address
	return err
}

func (c *Controller) Accept(options ...grpc.DialOption) (*grpc.ClientConn, error) {
	tcpconn, err := c.l.Accept()
	if err != nil {
		return nil, err
	}

	// save clientConn here for now so we can close it when our connection is dead, triggering a CANCELED code from grpc instead of
	// a connection error
	var clientConn *grpc.ClientConn
	wrappedConn := callbackCloser{
		Conn: tcpconn,
		callback: func() {
			if clientConn != nil {
				clientConn.Close()
			}
		},
		once: &sync.Once{},
	}

	once := &sync.Once{}

	// completely disregard context timeouts because there is always a connection ready. There is nothing to actually dial
	clientConn, err = grpc.Dial(
		// this is the listen address but that doesn't matter as the only usage is the
		// dialer which promptly ignores it
		c.addr,
		// apply our 'dialer' with the rest of the grpc options
		append(options, grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			var conn net.Conn
			var err = ErrWorkerDisconnected

			// ensure a connection is never reused
			once.Do(func() {
				conn = wrappedConn
				err = nil
			})

			return conn, err
		}))...)

	return clientConn, err
}
