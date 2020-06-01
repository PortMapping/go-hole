package lurker

import (
	"context"
	"fmt"
	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/common"
	"net"
	"net/http"

	"github.com/portmapping/lurker/nat"
)

type wsListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	port        int
	mappingPort int
	nat         nat.NAT
	listener    net.Listener
	cfg         *Config
	ready       bool
	handler     http.Handler
	certFile    string
	keyFile     string
}

// Listen ...
func (ws *wsListener) Listen(c chan<- Connector) (err error) {
	tcpAddr := common.LocalTCPAddr(ws.port)
	if ws.cfg.UseSecret {
		ws.listener, err = reuse.ListenTLS("tcp", DefaultLocalTCPAddr.String(), l.cfg.secret)
		http.Serve(ws.listener, ws.handler)
	} else {
		ws.listener, err = reuse.ListenTCP("tcp", tcpAddr)
		http.ServeTLS(ws.listener, ws.handler, ws.certFile, ws.keyFile)
	}
	if err != nil {
		return err
	}
	fmt.Println("listen tcp on address:", tcpAddr.String())
	ws.ready = true
	return
}

// Stop ...
func (ws *wsListener) Stop() error {
	panic("implement me")
}

// IsReady ...
func (ws *wsListener) IsReady() bool {
	panic("implement me")
}

// IsSupport ...
func (ws *wsListener) IsSupport() bool {
	panic("implement me")
}

// NAT ...
func (ws *wsListener) NAT() nat.NAT {
	panic("implement me")
}

func (ws *wsListener) newHandle() {

}
