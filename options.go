package lurker

import (
	"fmt"
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
func LocalAddr(network string) string {
	if strings.Index(network, "tcp") >= 0 {
		return DefaultLocalTCPAddr
	}
	return DefaultLocalUDPAddr
}
