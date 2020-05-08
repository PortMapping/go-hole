package lurker

import (
	"encoding/json"
	"fmt"
	"net"
)

// Source ...
type Source interface {
	net.Addr
	Ping() bool
	Decode(src interface{}) error
}

// Addr ...
type Addr struct {
	Network string
	Address string
}

type source struct {
	addr Addr
	data []byte
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

// Decode ...
func (c source) Decode(src interface{}) error {
	return json.Unmarshal(c.data, src)
}

// Ping ...
func (c source) Ping() bool {
	dial, err := net.Dial(c.Network(), c.String())
	if err != nil {
		return false
	}
	_, err = dial.Write([]byte("hello world"))
	if err != nil {
		return false
	}
	data := make([]byte, maxByteSize)
	read, err := dial.Read(data)
	if err != nil {
		return false
	}
	fmt.Println("read", string(read))
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
