package lurker

import (
	"net"
	"strconv"
)

// Addr ...
type Addr struct {
	Protocol string `json:"protocol"`
	IP       net.IP `json:"ip"`
	Port     int    `json:"port"`
}

// Network ...
func (addr Addr) Network() string {
	return addr.Protocol
}

// Network ...
func (addr Addr) String() string {
	return net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port))
}

// UDP ...
func (addr Addr) UDP() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   addr.IP,
		Port: addr.Port,
	}
}

// TCP ...
func (addr Addr) TCP() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   addr.IP,
		Port: addr.Port,
	}
}

// IsZero ...
func (addr Addr) IsZero() bool {
	return addr.Protocol == "" && addr.IP.Equal(net.IPv4zero) && addr.Port == 0
}

// ParseSourceAddr ...
func ParseSourceAddr(network string, ip net.IP, port int) *Addr {
	return &Addr{
		Protocol: network,
		IP:       ip,
		Port:     port,
	}
}

// ParseNetAddr ...
func ParseNetAddr(addr net.Addr) *Addr {
	ip, port := ParseAddr(addr.String())
	return &Addr{
		Protocol: addr.Network(),
		IP:       ip,
		Port:     port,
	}
}
