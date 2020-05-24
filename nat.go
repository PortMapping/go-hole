package lurker

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

type SupportType int

type Support2 uint64

// Support ...
type Support struct {
	List [NetworkSupportMax]bool
}

func (s *Support2) Add() {

}
