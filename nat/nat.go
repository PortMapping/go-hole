package nat

import (
	"net"
)

// NAT ...
type NAT interface {
	Mapping() (err error)
	ExtPort() int
	Port() int
	StopMapping() (err error)
	Remapping() (err error)
	GetExternalAddress() (addr net.IP, err error)
	GetDeviceAddress() (addr net.IP, err error)
	GetInternalAddress() (addr net.IP, err error)
}
