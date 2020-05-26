package proxy

import (
	"errors"
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
}

// New ...
func New(protocol string, auth Authenticate) (Proxy, error) {
	switch protocol {
	case Socks5:
		return newSocks5Proxy(auth)
	}
	return nil, errors.New("protocol was not supported ")
}
