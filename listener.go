package lurker

import "github.com/portmapping/lurker/nat"

// Listener ...
type Listener interface {
	Listen() (err error)
	Stop() error
	IsReady() bool
	Accept() <-chan Connector
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
