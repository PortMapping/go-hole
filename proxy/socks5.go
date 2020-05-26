package proxy

const (
	socks5Version = uint8(5)
)

type socks5 struct {
}

func newSocks5Proxy() (Proxy, error) {
	return &socks5{}, nil
}
