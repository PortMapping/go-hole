package lurker

import (
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

// SupportType ...
type SupportType uint64

// Support2 ...
type Support2 uint64

// Support ...
type Support struct {
	List [NetworkSupportMax]bool
}

// Add ...
func (s *SupportType) Add(t SupportType) {
	*s = (*s) | (t)
}

func mapping(network string, port int) (n nat.NAT, err error) {
	n, err = nat.FromLocal(network, port)
	if err != nil {
		log.Debugw("nat error", "error", err)
		if err == p2pnat.ErrNoNATFound {
			//fmt.Println("listen tcp on address:", tcpAddr.String())
		}
		return nil, err
	} else {
		extPort, err := n.Mapping()
		if err != nil {
			log.Debugw("nat mapping error", "error", err)
			return nil, err
		}
		l.mappingPort = extPort

		address, err := l.nat.GetExternalAddress()
		if err != nil {
			log.Debugw("get external address error", "error", err)
			l.cfg.NAT = false
			return nil
		}
		addr := ParseSourceAddr("tcp", address, extPort)
		fmt.Println("tcp mapping on address:", addr.String())
	}
}
