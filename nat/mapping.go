package nat

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-nat"
	"go.uber.org/atomic"
)

const description = "mapping_port"

// DefaultTimeOut ...
var DefaultTimeOut = 60 * time.Second

type natClient struct {
	stop     *atomic.Bool
	timeout  time.Duration
	nat      nat.NAT
	port     int
	protocol string
	extport  int
}

// Port ...
func (n *natClient) Port() int {
	return n.port
}

func defaultNAT() nat.NAT {
	n, err := nat.DiscoverGateway()
	if err != nil {
		panic(err)
	}
	return n
}

// FromLocal ...
func FromLocal(protocol string, port int) (nat NAT, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return &natClient{
		stop:     atomic.NewBool(false),
		nat:      defaultNAT(),
		timeout:  DefaultTimeOut,
		protocol: protocol,
		port:     port,
		extport:  0,
	}, nil
}

// New ...
func New(n nat.NAT, protocol string, port int) NAT {
	return &natClient{
		stop:     atomic.NewBool(false),
		nat:      n,
		timeout:  DefaultTimeOut,
		protocol: protocol,
		port:     port,
	}
}

// SetTimeOut ...
func (n *natClient) SetTimeOut(t int) {
	n.timeout = time.Duration(t)
}

// Mapping ...
func (n *natClient) Mapping() (err error) {
	n.stop.Store(false)
	n.extport, err = n.nat.AddPortMapping(n.protocol, n.port, description, n.timeout)
	if err != nil {
		return err
	}

	go func() {
		t := time.NewTicker(30 * time.Second)
		defer func() {
			t.Stop()
			if e := recover(); e != nil {
				fmt.Println("panic error:", e)
			}
		}()

		for {
			//check mapping every 30 second
			<-t.C
			if n.stop.Load() {
				return
			}
			_, err = n.nat.AddPortMapping(n.protocol, n.port, description, n.timeout)
			if err != nil {
				panic(err)
			}

		}
	}()

	return nil
}

// Remapping ...
func (n *natClient) Remapping() (err error) {
	if err := n.StopMapping(); err != nil {
		return err
	}
	return n.Mapping()
}

// StopMapping ...
func (n *natClient) StopMapping() (err error) {
	if n.nat != nil {
		n.stop.Store(true)
		if err := n.nat.DeletePortMapping("tcp", n.port); err != nil {
			return err
		}
	}
	return nil
}

// GetExternalAddress ...
func (n *natClient) GetExternalAddress() (addr net.IP, err error) {
	return n.nat.GetExternalAddress()
}

// ExtPort ...
func (n *natClient) ExtPort() int {
	return n.extport
}

// GetDeviceAddress ...
func (n *natClient) GetDeviceAddress() (addr net.IP, err error) {
	return n.nat.GetDeviceAddress()
}

// GetInternalAddress ...
func (n *natClient) GetInternalAddress() (addr net.IP, err error) {
	return n.nat.GetInternalAddress()
}

// GetNAT ...
func (n *natClient) GetNAT() nat.NAT {
	return n.nat
}
