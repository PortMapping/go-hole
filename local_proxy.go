package lurker

import (
	"context"
	"fmt"
	"net"

	"github.com/portmapping/lurker/proxy"
)

type localProxy struct {
	ctx    context.Context
	cancel context.CancelFunc
	config Proxy
	port   int
	local  proxy.Proxy
	ready  bool
}

// RegisterLocalProxy ...
func RegisterLocalProxy(l Lurker, cfg *Config) (err error) {
	for _, p := range cfg.Proxy {
		lp, err := proxy.New(p.Type, proxy.Auth{
			Name: p.Name,
			Pass: p.Pass,
		})
		if err != nil {
			return err
		}
		ctx, cFunc := context.WithCancel(context.TODO())
		l.RegisterListener("", &localProxy{
			ctx:    ctx,
			cancel: cFunc,
			config: p,
			port:   p.Port,
			local:  lp,
		})
	}

	return nil
}

// Listen ...
func (p *localProxy) Listen(c chan<- Connector) (err error) {
	lis, err := p.local.ListenPort(p.port)
	if err != nil {
		return err
	}
	fmt.Println("listen proxy on port:", p.port)
	go p.accept(lis)
	p.ready = true
	return
}

func (p *localProxy) accept(lis net.Listener) {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			conn, err := lis.Accept()
			if err != nil {
				log.Debugw("debug|getClientFromTCP|Accept", "error", err)
				continue
			}
			p.local.Monitor(conn)
		}
	}
}

// Stop ...
func (p *localProxy) Stop() error {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	return nil
}

// IsReady ...
func (p *localProxy) IsReady() bool {
	return p.ready
}
