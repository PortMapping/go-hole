package lurker

import (
	"encoding/json"
	"fmt"
	"github.com/portmapping/go-reuse"
	"net"
)

// Source ...
type Source interface {
	net.Addr
	MappingPort() int
	SetMappingPort(int)
	Ping(msg string) bool
	Decode(src interface{}) error
}

// Addr ...
type Addr struct {
	Network string
	Address string
}

type source struct {
	addr        Addr
	mappingPort int
	data        []byte
}

// NewSource ...
func NewSource(network, address string) Source {
	return &source{
		addr: Addr{
			Network: network,
			Address: address,
		},
	}
}

// Network ...
func (c source) Network() string {
	return c.addr.Network
}

// String ...
func (c source) String() string {
	return c.addr.Address
}

// MappingPort ...
func (c source) MappingPort() int {
	return c.mappingPort
}

// SetMappingPort ...
func (c *source) SetMappingPort(i int) {
	c.mappingPort = i
}

// Decode ...
func (c source) Decode(src interface{}) error {
	return json.Unmarshal(c.data, src)
}

// Ping ...
func (c source) Ping(msg string) bool {
	local := LocalAddr(LocalPort(c.Network(), c.mappingPort))
	remote := c.String()
	fmt.Println("local", local, "remote", remote, "network", c.Network())
	dial, err := reuse.Dial(c.Network(), local, remote)
	if err != nil {
		return false
	}
	_, err = dial.Write([]byte(msg))
	if err != nil {
		return false
	}
	data := make([]byte, maxByteSize)
	read, err := dial.Read(data)
	if err != nil {
		return false
	}
	fmt.Println("received: ", string(data[:read]))
	return true
}

// JSON ...
func (addr *Addr) JSON() []byte {
	marshal, err := json.Marshal(addr)
	if err != nil {
		return nil
	}
	return marshal
}
