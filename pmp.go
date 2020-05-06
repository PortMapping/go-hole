package natpmp

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-nat"
	"go.uber.org/atomic"
)

var DefaultTimeOut = 30

type pmpClient struct {
	timeout int
	nat     nat.NAT
	port    int
	stop    *atomic.Bool
}

func NewNatFromLocal(port int) (NAT, error) {
	nat, err := nat.DiscoverGateway()
	if err != nil {
		return nil, err
	}
	return &pmpClient{
		stop:    atomic.NewBool(false),
		nat:     nat,
		timeout: DefaultTimeOut,
		port:    port,
	}, nil
}

func (n *pmpClient) Mapping() (port int, err error) {
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

func (n *pmpClient) Remapping() (port int, err error) {
	n.StopMapping()
	n.stop.Store(false)
	return n.Mapping()
}

func (n *pmpClient) StopMapping() (err error) {
	if n.nat != nil {
		if err := n.nat.DeletePortMapping("tcp", n.port); err != nil {
			return err
		}
		n.stop.Store(true)
	}
	return nil
}

func (n *pmpClient) GetExternalAddress() (addr net.IP, err error) {
	return n.nat.GetExternalAddress()
}

//
//func (n *pmpClient) AddPortMapping(protocol string, externalPort, internalPort int) (mappedExternalPort int, err error) {
//	// Note order of port arguments is switched between our AddPortMapping and the client's AddPortMapping.
//	response, err := n.client.AddPortMapping(protocol, internalPort, externalPort, n.timeout)
//	if err != nil {
//		return
//	}
//	mappedExternalPort = int(response.MappedExternalPort)
//	return
//}

//func (n *pmpClient) DeletePortMapping(protocol string, externalPort, internalPort int) (err error) {
//	_, err = n.nat.AddPortMapping(protocol, internalPort, "", 0)
//	return
//}
