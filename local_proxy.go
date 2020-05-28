package lurker

import (
	"context"
	"fmt"
	"github.com/portmapping/lurker/nat"
	"net"

	"github.com/portmapping/lurker/proxy"
)

type localProxy struct {
	ctx      context.Context
	cancel   context.CancelFunc
	pCfg     Proxy
	port     int
	local    proxy.Proxy
	ready    bool
	protocol string
}

// RegisterLocalProxy ...
func RegisterLocalProxy(l Lurker, cfg *Config) (err error) {
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
			n, err = mapping("tcp", p.Port)
			if err != nil {
				return err
			}
		}

		lp, err := proxy.New(p.Type, n, a)
		if err != nil {
			return err
		}
		ctx, cFunc := context.WithCancel(context.TODO())

		l.RegisterListener("", &localProxy{
			ctx:    ctx,
			cancel: cFunc,
			pCfg:   p,
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
	fmt.Printf("listen %v proxy on port: %v", p.pCfg.Type, p.port)
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
			go p.local.Monitor(conn)
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
