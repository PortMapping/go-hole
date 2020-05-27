package proxy

import (
	"errors"
	"github.com/portmapping/lurker/nat"
	"net"
)

// Socks5 ...
const Socks5 = "socks5"

// HTTP ...
const HTTP = "http"

// HTTPS ...
const HTTPS = "https"

// Proxy ...
type Proxy interface {
	Monitor(conn net.Conn)
	ListenPort(port int) (net.Listener, error)
}

// New ...
func New(protocol string, n nat.NAT, auth Authenticate) (Proxy, error) {
	switch protocol {
	case Socks5:
		return newSocks5Proxy(n, auth)
	}
	return nil, errors.New("protocol was not supported")
}
