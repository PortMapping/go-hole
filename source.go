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
	IP      net.IP
	Port    int
}

type source struct {
	addr        Addr
	mappingPort int
	data        []byte
}

// NewSource ...
func NewSource(network string, ip net.IP, port int) Source {
	return &source{
		addr: Addr{
			Network: network,
			IP:      ip,
			Port:    port,
		},
	}
}

// Network ...
func (c source) Network() string {
	return c.addr.Network
}

// String ...
func (c source) String() string {
	return fmt.Sprintf("%s:%d", c.addr.IP, c.addr.Port)
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
	remote := c.String()
	local := LocalAddr(LocalPort(c.Network(), c.mappingPort))
	var dial net.Conn
	var err error
	if c.mappingPort != 0 {
		dial, err = reuse.Dial(c.Network(), local, remote)
	} else {
		dial, err = net.Dial(c.Network(), remote)
	}

	fmt.Println("local", local, "remote", remote, "network", c.Network(), "mapping", c.mappingPort)
	if err != nil {
		fmt.Println("debug|Ping|Dial", err)
		return false
	}
	_, err = dial.Write([]byte(msg))
	if err != nil {
		fmt.Println("debug|Ping|Write", err)
		return false
	}
	data := make([]byte, maxByteSize)
	for {
		read, err := dial.Read(data)
		if err != nil {
			fmt.Println("debug|Ping|Read", err)
			return false
		}
		fmt.Println("received: ", string(data[:read]))
	}
}

// JSON ...
func (addr *Addr) JSON() []byte {
	marshal, err := json.Marshal(addr)
	if err != nil {
		return nil
	}
	return marshal
}
