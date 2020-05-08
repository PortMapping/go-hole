package lurker

import (
	"fmt"
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

// DefaultLocalTCPAddr ...
var DefaultLocalTCPAddr = fmt.Sprintf("0.0.0.0:%d", DefaultTCP)

// DefaultLocalUDPAddr ...
var DefaultLocalUDPAddr = fmt.Sprintf("0.0.0.0:%d", DefaultUDP)

// LocalAddr ...
func LocalAddr(port int) string {
	return fmt.Sprintf("0.0.0.0:%d", port)
}

// LocalPort ...
func LocalPort(network string, mappingPort int) int {
	if strings.Index(network, "tcp") >= 0 {
		if mappingPort == 0 {
			return DefaultTCP
		}
		return mappingPort
	}
	return DefaultUDP
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
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("panic", e)
			return
		}
	}()
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
