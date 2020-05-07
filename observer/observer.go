package observer

import "github.com/PortMapping/go-hole"

// Observer ...
type Observer interface {
}

type observer struct {
	udpPort int
	tcpPort int
}

// New ...
func New() Observer {
	return &observer{
		udpPort: hole.DefaultUDP,
		tcpPort: hole.DefaultTCP,
	}
}
