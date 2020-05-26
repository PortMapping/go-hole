package lurker

import (
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Proxy ...
type Proxy struct {
	Type string `json:"type"`
	Port int    `json:"port"`
	Name string `json:"name"`
	Pass string `json:"pass"`
}

// Config ...
type Config struct {
	TCP         int
	UDP         int
	NAT         bool
	UseProxy    bool
	Proxy       []Proxy
	UseSecret   bool
	Certificate string
	secret      *tls.Config
}

// DefaultTimeout ...
var DefaultTimeout = 60 * time.Second

// DefaultConnectionTimeout ...
var DefaultConnectionTimeout = 15 * time.Second

// DefaultTCP ...
var DefaultTCP = 46666

// DefaultUDP ...
var DefaultUDP = 47777

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

// GlobalID ...
var GlobalID string

func init() {
	GlobalID = UUID()
}

// UUID ...
func UUID() string {
	return uuid.Must(uuid.NewUUID()).String()
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		TCP:      DefaultTCP,
		UDP:      DefaultUDP,
		NAT:      true,
		UseProxy: true,
		Proxy: []Proxy{
			{
				Type: "socks5",
				Port: 10080,
				Name: "",
				Pass: "",
			},
		},
		UseSecret:   false,
		Certificate: "",
		secret:      nil,
	}
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

// TCPAddr ...
func TCPAddr(ip net.IP, port int) *net.TCPAddr {
	return &net.TCPAddr{
		IP:   ip,
		Port: port,
	}
}

// ParseTCPAddr ...
func ParseTCPAddr(addr string) *net.TCPAddr {
	ip, port := ParseAddr(addr)
	return &net.TCPAddr{
		IP:   ip,
		Port: port,
	}
}

// UDPAddr ...
func UDPAddr(ip net.IP, port int) *net.UDPAddr {
	return &net.UDPAddr{
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
