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

func copyConnGroup(group interface{}) {
	cg, ok := group.(connGroup)
	if !ok {
		return
	}
	var err error
	*cg.n, err = io.Copy(cg.dst, cg.src)
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
	return p
}

// AddConnections ...
func (p *pool) AddConnections(conn Connection) {
	p.pool.Submit(func() {

	})
}
func (p *pool) connectsForward(group interface{}) {
	conns := group.(Connection)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	var in, out int64
	_ = p.copyPool.Invoke(newConnGroup(conns.conn1, conns.conn2, wg, &in))
	// outside to mux : incoming
	_ = p.copyPool.Invoke(newConnGroup(conns.conn2, conns.conn1, wg, &out))
	// mux to outside : outgoing
	fmt.Println("in", in, "out", out)
	wg.Wait()
	conns.wg.Done()
}

// AddConnections ...
func AddConnections(conn Connection) {

}
