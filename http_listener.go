package lurker

import (
	"context"
	"fmt"
	p2pnat "github.com/libp2p/go-nat"
	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/nat"
	"net"
	"net/http"
)

type httpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	port        int
	mappingPort int
	nat         nat.NAT
	tcpListener net.Listener
	cfg         *Config
	handler     http.Handler
	srv         *http.Server
}

// MappingPort ...
func (l *httpListener) MappingPort() int {
	return l.mappingPort
}

// Listen ...
func (l *httpListener) Listen(c chan<- Source) (err error) {
	tcpAddr := LocalTCPAddr(l.port)
	if l.cfg.Secret != nil {
		l.tcpListener, err = reuse.ListenTLS("tcp", DefaultLocalTCPAddr.String(), l.cfg.Secret)
	} else {
		l.tcpListener, err = reuse.ListenTCP("tcp", tcpAddr)
	}
	if err != nil {
		return err
	}
	l.srv = &http.Server{Handler: l.handler}
	fmt.Println("listen http on address:", tcpAddr.String())
	go listenHTTP(l.ctx, l.srv, l.tcpListener, c)
	if !l.cfg.NAT {
		return nil
	}

	l.nat, err = nat.FromLocal("tcp", l.cfg.TCP)
	if err != nil {
		log.Debugw("nat error", "error", err)
		if err == p2pnat.ErrNoNATFound {
			//fmt.Println("listen tcp on address:", tcpAddr.String())
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
		fmt.Println("http mapping on address:", addr.String())
	}
	return
}

// Stop ...
func (l *httpListener) Stop() error {
	if l.srv != nil {
		return l.srv.Close()
	}
	return nil
}

// NewHTTPListener ...
func NewHTTPListener(cfg *Config, handler http.Handler) Listener {
	h := &httpListener{
		ctx:     nil,
		cancel:  nil,
		handler: handler,
		port:    cfg.HTTP,
		cfg:     cfg,
	}
	h.ctx, h.cancel = context.WithCancel(context.TODO())
	return h
}
func listenHTTP(ctx context.Context, srv *http.Server, l net.Listener, s chan<- Source) {
	err := srv.Serve(l)
	if err != nil {
		return
	}
}
