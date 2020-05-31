package lurker

import (
	"context"
	"fmt"
	"net"

	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/common"
	"github.com/portmapping/lurker/nat"
)

type tcpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
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
	go listenTCP(l.ctx, l.listener, c)
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

func listenTCP(ctx context.Context, listener net.Listener, cli chan<- Connector) (err error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Debugw("debug|getClientFromTCP|Accept", "error", err)
				continue
			}
			fmt.Println("new connector")
			go tcpProcess(conn, cli)
			return nil
		}
	}
}

func tcpProcess(conn net.Conn, cli chan<- Connector) {
	t := newTCPConnector(conn)
	t.Process()
	cli <- t
}
