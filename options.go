package lurker

import (
	"net"
	"strconv"
	"strings"
)

// DefaultTimeout ...
var DefaultTimeout = 60

// DefaultTCP ...
var DefaultTCP = 46666

// DefaultUDP ...
var DefaultUDP = 47777

// DefaultHolePort ...
var DefaultHolePort = 0

// DefaultLocalTCPAddr ...
var DefaultLocalTCPAddr = &net.TCPAddr{
	IP:   net.IPv4zero,
	Port: DefaultTCP,
}

// DefaultLocalUDPAddr ...
var DefaultLocalUDPAddr = &net.UDPAddr{
	IP:   net.IPv4zero,
	Port: DefaultUDP,
}

// LocalUDPAddr ...
func LocalUDPAddr(port int) *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}
}

// LocalTCPAddr ...
func LocalTCPAddr(port int) *net.TCPAddr {
	return &net.TCPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}
}

// LocalAddr ...
func LocalAddr(ip net.IP, port int) string {
	return net.JoinHostPort(ip.String(), strconv.Itoa(port))
}

// ParseTCPAddr ...
func ParseTCPAddr(addr string) *net.TCPAddr {
	ip, port := ParseAddr(addr)
	return &net.TCPAddr{
		IP:   ip,
		Port: port,
	}
}

// ParseUDPAddr ...
func ParseUDPAddr(addr string) *net.UDPAddr {
	ip, port := ParseAddr(addr)
	return &net.UDPAddr{
		IP:   ip,
		Port: port,
	}
}

// ParseAddr ...
func ParseAddr(addr string) (net.IP, int) {
	addrs := strings.Split(addr, ":")
	ip := net.ParseIP(addrs[0])
	if len(addrs) > 1 {
		port, err := strconv.ParseInt(addrs[1], 10, 32)
		if err != nil {
			return ip, 0
		}
		return ip, int(port)
	}
	return ip, 0
}

// IsUDP ...
func IsUDP(network string) bool {
	if strings.Index(network, "udp") >= 0 {
		return true
	}
	return false
}

// IsTCP ...
func IsTCP(network string) bool {
	if strings.Index(network, "tcp") >= 0 {
		return true
	}
	return false
}
