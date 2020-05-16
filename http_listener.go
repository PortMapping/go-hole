package lurker

import (
	"context"
	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/nat"
	"net"
	"net/http"
)

type httpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	source      chan Source
	port        int
	mappingPort int
	nat         nat.NAT
	tcpListener net.Listener
	cfg         *Config
	handler     http.Handler
}

// Listen ...
func (l *httpListener) Listen() (c <-chan Source, err error) {
	tcpAddr := LocalTCPAddr(l.port)
	if l.cfg.Secret != nil {
		l.tcpListener, err = reuse.ListenTLS("tcp", DefaultLocalTCPAddr.String(), l.cfg.Secret)
	} else {
		l.tcpListener, err = reuse.ListenTCP("tcp", tcpAddr)
	}
	if err != nil {
		return nil, err
	}
	go listenHTTP(l.ctx, l.tcpListener, l.handler, l.source)

	return
}

// Stop ...
func (l *httpListener) Stop() error {
	panic("implement me")
}

// NewHTTPListener ...
func NewHTTPListener(cfg *Config, handler http.Handler) Listener {
	h := &httpListener{
		ctx:     nil,
		cancel:  nil,
		handler: handler,
		port:    cfg.HTTP,
		cfg:     cfg,
		source:  make(chan Source),
	}
	h.ctx, h.cancel = context.WithCancel(context.TODO())
	return h
}
func listenHTTP(ctx context.Context, l net.Listener, handler http.Handler, s chan<- Source) {
	srv := &http.Server{Handler: handler}
	err := srv.Serve(l)
	if err != nil {
		return
	}
}
