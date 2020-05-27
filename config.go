package lurker

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/google/uuid"
)

// Proxy ...
type Proxy struct {
	Type string `json:"type"`
	Nat  bool   `json:"nat"`
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
