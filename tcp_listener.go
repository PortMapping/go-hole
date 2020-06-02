package lurker

import (
	"context"
	"fmt"
	"net"

	"github.com/panjf2000/ants/v2"
	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/common"
	"github.com/portmapping/lurker/nat"
)

type tcpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	funcPool    *ants.PoolWithFunc
	port        int
	mappingPort int
	nat         nat.NAT
	listener    net.Listener
	cfg         *Config
	ready       bool
}

// IsSupport ...
func (l *tcpListener) IsSupport() bool {
	return l.cfg.NAT && l.nat != nil
}

// NAT ...
func (l *tcpListener) NAT() nat.NAT {
	return l.nat
}

// IsReady ...
func (l *tcpListener) IsReady() bool {
	return l.ready
}

// MappingPort ...
func (l *tcpListener) MappingPort() int {
	return l.mappingPort
}

// NewTCPListener ...
func NewTCPListener(cfg *Config) Listener {
	tcp := &tcpListener{
		ctx:    nil,
		cancel: nil,
		port:   cfg.TCP,
		cfg:    cfg,
	}
	tcp.ctx, tcp.cancel = context.WithCancel(context.TODO())
	var err error
	if cfg.NAT {
		tcp.nat, err = mapping("tcp", cfg.TCP)
		if err != nil {
			panic(err)
		}
	}
	tcp.funcPool, err = ants.NewPoolWithFunc(ants.DefaultAntsPoolSize, tcpHandler, func(opts *ants.Options) {
		opts.Nonblocking = false
	})
	return tcp
}

// Listen ...
func (l *tcpListener) Listen(c chan<- Connector) (err error) {
	tcpAddr := common.LocalTCPAddr(l.port)
	if l.cfg.UseSecret {
		l.listener, err = reuse.ListenTLS("tcp", DefaultLocalTCPAddr.String(), l.cfg.secret)
	} else {
		l.listener, err = reuse.ListenTCP("tcp", tcpAddr)
	}
	if err != nil {
		return err
	}
	fmt.Println("listen tcp on address:", tcpAddr.String())
	go l.listenTCP(c)
	l.ready = true
	return
}

// Stop ...
func (l *tcpListener) Stop() error {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	return nil
}

func (l *tcpListener) listenTCP(c chan<- Connector) (err error) {
	for {
		select {
		case <-l.ctx.Done():
			return
		default:
			conn, err := l.listener.Accept()
			if err != nil {
				log.Debugw("debug|getClientFromTCP|Accept", "error", err)
				continue
			}
			fmt.Println("new connector")
			t := newTCPConnector(conn)
			err = l.funcPool.Invoke(t)
			if err != nil {
				log.Debugw("debug|funcPool|Invoke", "error", err)
				continue
			}
			c <- t
			return nil
		}
	}
}

func tcpHandler(i interface{}) {
	connector, b := i.(Connector)
	if !b {
		return
	}
	connector.Processing()
}
