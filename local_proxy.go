package lurker

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/portmapping/lurker/nat"
	"net"

	"github.com/portmapping/lurker/proxy"
)

type localProxy struct {
	ctx      context.Context
	cancel   context.CancelFunc
	proxyCfg Proxy
	port     int
	local    proxy.Proxy
	ready    bool
	protocol string
	funcPool *ants.PoolWithFunc
}

// RegisterLocalProxy ...
func RegisterLocalProxy(l Lurker, cfg *Config) (port int, err error) {
	for _, p := range cfg.Proxy {
		a := proxy.NoAuth()
		if p.Name != "" && p.Pass != "" {
			a = proxy.Auth{
				Name: p.Name,
				Pass: p.Pass,
			}
		}
		var n nat.NAT
		if p.Nat {
			//todo(network can change)
			n, err = Mapping("tcp", p.Port)
			if err != nil {
				return 0, err
			}
			port = n.ExtPort()
		}

		lp, err := proxy.New(p.Type, n, a)
		if err != nil {
			return 0, err
		}
		ctx, cFunc := context.WithCancel(context.TODO())

		l.RegisterListener("", &localProxy{
			ctx:      ctx,
			cancel:   cFunc,
			proxyCfg: p,
			port:     p.Port,
			local:    lp,
		})
	}

	return port, nil
}

// Listen ...
func (p *localProxy) Listen(c chan<- Connector) (err error) {
	lis, err := p.local.ListenOnPort(p.port)
	if err != nil {
		return err
	}
	log.Infof("listen %v proxy on port: %v", p.proxyCfg.Type, p.port)
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
			fmt.Println("new connect received")
			err = p.local.Connect(conn)
			if err != nil {
				log.Debugw("debug|Connect|error", "error", err)
				continue
			}
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
