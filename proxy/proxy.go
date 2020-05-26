package proxy

import "errors"

// Socks5 ...
const Socks5 = "socks5"

// HTTP ...
const HTTP = "http"

// HTTPS ...
const HTTPS = "https"

// Proxy ...
type Proxy interface {
}

// New ...
func New(protocol string) (Proxy, error) {
	switch protocol {
	case Socks5:
		return newSocks5Proxy()
	}
	return nil, errors.New("protocol was not supported ")
}
