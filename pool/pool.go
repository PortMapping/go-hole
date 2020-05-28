package pool

import (
	"fmt"
	"io"
	"sync"

	"github.com/panjf2000/ants/v2"
)

type connGroup struct {
	src io.ReadWriteCloser
	dst io.ReadWriteCloser
	wg  *sync.WaitGroup
	n   *int64
}

// Connection ...
type Connection struct {
	conn1 io.ReadWriteCloser
	conn2 io.ReadWriteCloser
	wg    *sync.WaitGroup
}

type pool struct {
	copyPool *ants.PoolWithFunc
	pool     *ants.Pool
}

// Pool ...
type Pool interface {
	AddConnections(conn Connection)
}

var defaultPool Pool

func init() {
	defaultPool = NewPool()
}

func newConnGroup(dst, src io.ReadWriteCloser, wg *sync.WaitGroup, n *int64) connGroup {
	return connGroup{
		src: src,
		dst: dst,
		wg:  wg,
		n:   n,
	}
}

func doCopy(src, dst io.ReadWriteCloser) (n int64, err error) {
	buf := make([]byte, 0xff)
	n1 := 0
	for {
		n1, err = src.Read(buf[0:])
		n = int64(n1)
		if err != nil {
			return
		}
		b := buf[0:n]
		_, err = dst.Write(b)
		if err != nil {
			return
		}
	}
}

func copyConnGroup(group interface{}) {
	cg, ok := group.(connGroup)
	if !ok {
		return
	}
	var err error
	*cg.n, err = doCopy(cg.dst, cg.src)
	if err != nil {
		cg.src.Close()
		cg.dst.Close()
	}
	cg.wg.Done()
}

// NewConnection ...
func NewConnection(conn1, conn2 io.ReadWriteCloser, wg *sync.WaitGroup) Connection {
	return Connection{
		conn1: conn1,
		conn2: conn2,
		wg:    wg,
	}
}

// NewPool ...
func NewPool() Pool {
	np, err := ants.NewPool(1000)
	if err != nil {
		panic(err)
	}

	fp, err := ants.NewPoolWithFunc(1000, copyConnGroup, ants.WithNonblocking(false))
	if err != nil {
		panic(err)
	}
	var p pool
	p.pool = np
	p.copyPool = fp
	return &p
}

// AddConnections ...
func (p *pool) AddConnections(conn Connection) {
	p.pool.Submit(func() {
		p.connectsForward(conn)
		conn.wg.Done()
	})
}
func (p *pool) connectsForward(c Connection) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	var in, out int64
	_ = p.copyPool.Invoke(newConnGroup(c.conn1, c.conn2, wg, &in))
	// outside to mux : incoming
	_ = p.copyPool.Invoke(newConnGroup(c.conn2, c.conn1, wg, &out))
	// mux to outside : outgoing
	wg.Wait()
	fmt.Println("in", in, "out", out)

}

// AddConnections ...
func AddConnections(conn Connection) {
	defaultPool.AddConnections(conn)
}
