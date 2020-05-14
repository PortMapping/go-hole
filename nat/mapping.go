package nat

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-nat"
	"go.uber.org/atomic"
)

// DefaultTimeOut ...
var DefaultTimeOut time.Duration = 60

type natClient struct {
	stop    *atomic.Bool
	timeout time.Duration
	nat     nat.NAT
	port    int
}

func defaultNAT() nat.NAT {
	n, err := nat.DiscoverGateway()
	if err != nil {
		panic(err)
	}
	return n
}

// FromLocal ...
func FromLocal(port int) (nat NAT, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return &natClient{
		stop:    atomic.NewBool(false),
		nat:     defaultNAT(),
		timeout: DefaultTimeOut,
		port:    port,
	}, nil
}

// New ...
func New(n nat.NAT, port int) NAT {
	return &natClient{
		stop:    atomic.NewBool(false),
		nat:     n,
		timeout: DefaultTimeOut,
		port:    port,
	}
}

// SetTimeOut ...
func (n *natClient) SetTimeOut(t int) {
	n.timeout = time.Duration(t)
}

// Mapping ...
func (n *natClient) Mapping() (port int, err error) {
	n.stop.Store(false)
	eport, err := n.nat.AddPortMapping("tcp", n.port, "mapping", n.timeout)
	if err != nil {
		return 0, err
	}
	port = eport

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
			_, err = n.nat.AddPortMapping("tcp", n.port, "http", n.timeout)
			if err != nil {
				panic(err)
			}

		}
	}()

	return port, nil
}

// Remapping ...
func (n *natClient) Remapping() (port int, err error) {
	if err := n.StopMapping(); err != nil {
		return 0, err
	}
	return n.Mapping()
}

// StopMapping ...
func (n *natClient) StopMapping() (err error) {
	if n.nat != nil {
		if err := n.nat.DeletePortMapping("tcp", n.port); err != nil {
			return err
		}
		n.stop.Store(true)
	}
	return nil
}

// GetExternalAddress ...
func (n *natClient) GetExternalAddress() (addr net.IP, err error) {
	return n.nat.GetExternalAddress()
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
