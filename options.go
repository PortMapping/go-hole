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
func LocalAddr(port int) string {
	return fmt.Sprintf("0.0.0.0:%d", port)
}

// LocalPort ...
func LocalPort(network string, mappingPort int) int {
	if mappingPort == 0 {
		if strings.Index(network, "tcp") >= 0 {
			return DefaultTCP
		}
		return DefaultUDP
	}
	return mappingPort
}
