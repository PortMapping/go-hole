package natpmp

import "net"

type NAT interface {
	GetExternalAddress() (addr net.IP, err error)
	AddPortMapping(protocol string, externalPort, internalPort int) (mappedExternalPort int, err error)
	DeletePortMapping(protocol string, externalPort, internalPort int) (err error)
}
