package lurker

import (
	"fmt"
	p2pnat "github.com/libp2p/go-nat"
	"github.com/portmapping/lurker/nat"
)

// PublicNetworkTCP ...
const (
	PublicNetworkTCP = iota
	PublicNetworkUDP
	ProviderNetworkTCP
	ProviderNetworkUDP
	PrivateNetworkTCP
	PrivateNetworkUDP
	NetworkSupportMax
)

// PublicNetworkTCP ...
const (
	SupportTypePubliccTCP SupportType = 1 << iota
	SupportTypePublicUDP
	SupportTypeProviderTCP
	SupportTypeProviderUDP
	SupportTypePrivateTCP
	SupportTypePrivateUDP
	SupportTypeMax
)

// SupportType ...
type SupportType uint64

// Support ...
type Support struct {
	List [NetworkSupportMax]bool
	Type SupportType
}

// NATer ...
type NATer interface {
	IsSupport() bool
	NAT() nat.NAT
}

// Add ...
func (s *SupportType) Add(t SupportType) {
	*s = (*s) | (t)
}

// Del ...
func (s *SupportType) Del(t SupportType) {
	*s = (*s) ^ (t)
}

func mapping(network string, port int) (n nat.NAT, err error) {
	n, err = nat.FromLocal(network, port)
	if err != nil {
		log.Debugw("nat error", "error", err)
		if err == p2pnat.ErrNoNATFound {
			//fmt.Println("listen tcp on address:", tcpAddr.String())
		}
		return nil, err
	}
	err = n.Mapping()
	if err != nil {
		log.Debugw("nat mapping error", "error", err)
		return nil, err
	}

	address, err := n.GetExternalAddress()
	if err != nil {
		log.Debugw("get external address error", "error", err)
		return nil, err
	}
	addr := ParseSourceAddr("tcp", address, n.ExtPort())
	fmt.Printf("%s mapping on address: %v\n", network, addr)
	return n, nil
}
