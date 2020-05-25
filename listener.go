package lurker

import "github.com/portmapping/lurker/nat"

// Listener ...
type Listener interface {
	Listen(c <-chan Connector) (err error)
	Stop() error
	IsReady() bool
}

// PortMapping ...
type PortMapping interface {
	IsSupport() bool
	NAT() nat.NAT
}

// MappingListener ...
type MappingListener interface {
	Listener
	PortMapping
}
