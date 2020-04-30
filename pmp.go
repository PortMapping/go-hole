package natpmp

import (
	"github.com/jackpal/gateway"
	natpmp "github.com/jackpal/go-nat-pmp"
	"net"
)

var DefaultTimeOut = 30

type pmpClient struct {
	timeout int
	client  *natpmp.Client
}

func NewNatFromLocal() (NAT, error) {
	gatewayIP, err := gateway.DiscoverGateway()
	if err != nil {
		return nil, err
	}
	return &pmpClient{
		client:  natpmp.NewClient(gatewayIP),
		timeout: DefaultTimeOut,
	}, nil
}

func (n *pmpClient) GetExternalAddress() (addr net.IP, err error) {
	response, err := n.client.GetExternalAddress()
	if err != nil {
		return
	}
	ip := response.ExternalIPAddress
	addr = net.IPv4(ip[0], ip[1], ip[2], ip[3])
	return
}

func (n *pmpClient) AddPortMapping(protocol string, externalPort, internalPort int) (mappedExternalPort int, err error) {
	// Note order of port arguments is switched between our AddPortMapping and the client's AddPortMapping.
	response, err := n.client.AddPortMapping(protocol, internalPort, externalPort, n.timeout)
	if err != nil {
		return
	}
	mappedExternalPort = int(response.MappedExternalPort)
	return
}

func (n *pmpClient) DeletePortMapping(protocol string, externalPort, internalPort int) (err error) {
	_, err = n.client.AddPortMapping(protocol, internalPort, 0, 0)
	return
}
