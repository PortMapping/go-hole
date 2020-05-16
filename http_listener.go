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
	srv         *http.Server
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
	l.srv = &http.Server{Handler: l.handler}
	go listenHTTP(l.ctx, l.srv, l.tcpListener, l.source)

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
func listenHTTP(ctx context.Context, srv *http.Server, l net.Listener, s chan<- Source) {
	err := srv.Serve(l)
	if err != nil {
		return
	}
}
