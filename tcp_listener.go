package lurker

import (
	"context"
	"fmt"
	"github.com/portmapping/go-reuse"
	"net"

	p2pnat "github.com/libp2p/go-nat"
	"github.com/portmapping/lurker/nat"
)

type tcpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	port        int
	mappingPort int
	nat         nat.NAT
	tcpListener net.Listener
	cfg         *Config
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
	return tcp
}

// Listen ...
func (l *tcpListener) Listen(c chan<- Source) (err error) {
	tcpAddr := LocalTCPAddr(l.port)
	if l.cfg.Secret != nil {
		l.tcpListener, err = reuse.ListenTLS("tcp", DefaultLocalTCPAddr.String(), l.cfg.Secret)
	} else {
		l.tcpListener, err = reuse.ListenTCP("tcp", tcpAddr)
	}
	if err != nil {
		return err
	}
	go listenTCP(l.ctx, l.tcpListener, c)

	if !l.cfg.NAT {
		return nil
	}

	l.nat, err = nat.FromLocal(l.cfg.TCP)
	if err != nil {
		log.Debugw("nat error", "error", err)
		if err == p2pnat.ErrNoNATFound {
			fmt.Println("listen tcp on address:", tcpAddr.String())
		}
		l.cfg.NAT = false
	} else {
		extPort, err := l.nat.Mapping()
		if err != nil {
			log.Debugw("nat mapping error", "error", err)
			l.cfg.NAT = false
			return nil
		}
		l.mappingPort = extPort

		address, err := l.nat.GetExternalAddress()
		if err != nil {
			log.Debugw("get external address error", "error", err)
			l.cfg.NAT = false
			return nil
		}
		addr := ParseSourceAddr("tcp", address, extPort)
		fmt.Println("mapping on address:", addr.String())
	}
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
