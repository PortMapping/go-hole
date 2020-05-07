package nat

import (
	"github.com/libp2p/go-nat"
	"net"
)

// NAT ...
type NAT interface {
	Mapping() (port int, err error)
	StopMapping() (err error)
	Remapping() (port int, err error)
	GetExternalAddress() (addr net.IP, err error)
	GetDeviceAddress() (addr net.IP, err error)
	GetInternalAddress() (addr net.IP, err error)
	GetNAT() nat.NAT
}
