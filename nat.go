package lurker

import (
	"fmt"
	p2pnat "github.com/libp2p/go-nat"
	address2 "github.com/portmapping/lurker/common"
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

// Mapping ...
func Mapping(network string, port int) (n nat.NAT, err error) {
	n, err = nat.FromLocal(network, port)
	if err != nil {
		log.Debugw("nat error", "error", err)
		if err == p2pnat.ErrNoNATFound {
			//fmt.Println("listen tcp on common:", tcpAddr.String())
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
		log.Debugw("get external common error", "error", err)
		return nil, err
	}
	addr := address2.ParseSourceAddr("tcp", address, n.ExtPort())
	fmt.Printf("%s listen port %v was mapping on address: %v\n", network, port, addr)
	return n, nil
}
