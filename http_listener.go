package lurker

import (
	"context"
	"github.com/portmapping/lurker/nat"
	"net"
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
}

// Listen ...
func (l *httpListener) Listen() (c <-chan Source, err error) {
	return
}

// Stop ...
func (l *httpListener) Stop() error {
	panic("implement me")
}

// NewHTTPListener ...
func NewHTTPListener(cfg *Config) Listener {
	h := &httpListener{
		ctx:    nil,
		cancel: nil,
		port:   cfg.HTTP,
		cfg:    cfg,
		source: make(chan Source),
	}
	h.ctx, h.cancel = context.WithCancel(context.TODO())
	return h
}
