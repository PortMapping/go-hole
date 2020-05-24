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

type SupportType uint64

type Support2 uint64

// Support ...
type Support struct {
	List [NetworkSupportMax]bool
}

func (s *SupportType) Add(t SupportType) {
	*s = (*s) | (t)
}
