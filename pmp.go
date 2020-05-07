package hole

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-nat"
	"go.uber.org/atomic"
)

var DefaultTimeOut = 30

type natClient struct {
	stop    *atomic.Bool
	timeout int
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

func NewNATFromLocal(port int) (nat NAT, err error) {
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

func NewNAT(n nat.NAT, port int) NAT {
	return &natClient{
		stop:    atomic.NewBool(false),
		nat:     n,
		timeout: DefaultTimeOut,
		port:    port,
	}
}

func (n *natClient) SetTimeOut(t int) {
	n.timeout = t
}

func (n *natClient) Mapping() (port int, err error) {
	eport, err := n.nat.AddPortMapping("tcp", n.port, "http", 60)
	if err != nil {
		return 0, err
	}
	port = eport
	go func() {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println("panic error:", e)
				//err = e.(error)
			}
		}()

		for {
			time.Sleep(30 * time.Second)
			_, err = n.nat.AddPortMapping("tcp", n.port, "http", 60)
			if err != nil {
				panic(err)
			}
			if n.stop.Load() {
				return
			}
		}
	}()

	return port, nil
}

func (n *natClient) Remapping() (port int, err error) {
	n.StopMapping()
	n.stop.Store(false)
	return n.Mapping()
}

func (n *natClient) StopMapping() (err error) {
	if n.nat != nil {
		if err := n.nat.DeletePortMapping("tcp", n.port); err != nil {
			return err
		}
		n.stop.Store(true)
	}
	return nil
}

func (n *natClient) GetExternalAddress() (addr net.IP, err error) {
	return n.nat.GetExternalAddress()
}

func (n *natClient) GetDeviceAddress() (addr net.IP, err error) {
	return n.nat.GetDeviceAddress()
}

func (n *natClient) GetInternalAddress() (addr net.IP, err error) {
	return n.nat.GetInternalAddress()
}

func (n *natClient) GetNAT() nat.NAT {
	return n.nat
}

//
//func (n *natClient) AddPortMapping(protocol string, externalPort, internalPort int) (mappedExternalPort int, err error) {
//	// Note order of port arguments is switched between our AddPortMapping and the client's AddPortMapping.
//	response, err := n.client.AddPortMapping(protocol, internalPort, externalPort, n.timeout)
//	if err != nil {
//		return
//	}
//	mappedExternalPort = int(response.MappedExternalPort)
//	return
//}

//func (n *natClient) DeletePortMapping(protocol string, externalPort, internalPort int) (err error) {
//	_, err = n.nat.AddPortMapping(protocol, internalPort, "", 0)
//	return
//}
