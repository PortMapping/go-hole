package lurker

import (
	"github.com/portmapping/lurker/common"
	"net"
	"testing"
)

// TestSource_Connect ...
func TestSource_Connect(t *testing.T) {
	s := Service{
		ID:          GlobalID,
		Addr:        nil,
		ISP:         nil,
		Local:       nil,
		PortUDP:     0,
		PortTCP:     0,
		KeepConnect: false,
	}
	ip := net.ParseIP("172.0.0.1")
	port := 16004
	ss := NewSource(s, common.Addr{
		Protocol: "tcp",
		IP:       ip,
		Port:     port,
	})
	if err := ss.Connect(); err != nil {
		t.Fatal(err)
	}
}
