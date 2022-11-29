package reversegrpc

import (
	"context"
	"net"
	"sync"

	"golang.org/x/sync/semaphore"
)

// who needs to listen anyways. just dial :svekw:
type DialListener struct {
	sema *semaphore.Weighted
	addr string

	// laddr contains the local address of the current connection. This is the
	// closest we can get to the listening address of a 'server'
	laddr net.Addr
}

func NewDialListener(addr string) *DialListener {
	return &DialListener{
		sema: semaphore.NewWeighted(1),
		addr: addr,
	}
}

// Accepts dials out the controller and returns the tcp connection
// that has been set up. A semaphore is used to ensure there is a most
// one live connection
func (d *DialListener) Accept() (net.Conn, error) {
	d.sema.Acquire(context.Background(), 1)

	conn, err := net.Dial("tcp", d.addr)
	if err == nil {
		d.laddr = conn.LocalAddr()
		return callbackCloser{
			Conn: conn,
			callback: func() {
				d.sema.Release(1)
			},
			once: &sync.Once{},
		}, nil
	}

	return nil, err
}

// nothing to close when you aren't listening
func (d *DialListener) Close() error {
	return nil
}

func (d DialListener) Addr() net.Addr {
	return d.laddr
}

// callbackCloser performs a callback when Close is called
type callbackCloser struct {
	net.Conn
	callback func()
	once     *sync.Once
}

func (c callbackCloser) Close() error {
	c.once.Do(c.callback)
	return c.Conn.Close()
}
