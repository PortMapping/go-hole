package natpmp

import (
	"log"
	"time"

	"github.com/libp2p/go-nat"
	"go.uber.org/atomic"
)

var DefaultTimeOut = 30

type pmpClient struct {
	timeout int
	//client  *natpmp.Client
	nat  nat.NAT
	port int
	stop *atomic.Bool
}

func NewNatFromLocal(port int) (NAT, error) {
	//gatewayIP, err := gateway.DiscoverGateway()
	//if err != nil {
	//	return nil, err
	//}
	nat, err := nat.DiscoverGateway()
	if err != nil {
		return nil, err
	}
	return &pmpClient{
		//client:  natpmp.NewClient(gatewayIP),
		stop:    atomic.NewBool(false),
		nat:     nat,
		timeout: DefaultTimeOut,
		port:    port,
	}, nil
}

func (n *pmpClient) Mapping() (port int, err error) {

	log.Printf("nat type: %s", n.nat.Type())
	daddr, err := n.nat.GetDeviceAddress()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("device address: %s", daddr)

	iaddr, err := n.nat.GetInternalAddress()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("internal address: %s", iaddr)

	eaddr, err := n.nat.GetExternalAddress()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("external address: %s", eaddr)

	eport, err := n.nat.AddPortMapping("tcp", n.port, "http", 60)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	port = eport
	log.Printf("test-page: http://%s:%d/", eaddr, eport)
	go func() {
		for {
			time.Sleep(30 * time.Second)
			_, err = n.nat.AddPortMapping("tcp", n.port, "http", 60)
			if err != nil {
				log.Fatalf("error: %s", err)
			}
			if n.stop.Load() {
				return
			}
		}
	}()
	//defer nat.DeletePortMapping("tcp", 16005)

	return port, nil
}

func (n *pmpClient) StopMapping() {
	if n.nat != nil {
		n.nat.DeletePortMapping("tcp", n.port)
		n.stop.Store(true)
	}
}

//func (n *pmpClient) GetExternalAddress() (addr net.IP, err error) {
//	response, err := n.client.GetExternalAddress()
//	if err != nil {
//		return
//	}
//	ip := response.ExternalIPAddress
//	addr = net.IPv4(ip[0], ip[1], ip[2], ip[3])
//	return
//}
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
