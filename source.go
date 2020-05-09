package lurker

import (
	"encoding/json"
	"fmt"
	"github.com/portmapping/go-reuse"
	"net"
	"strconv"
	"time"
)

// Source ...
type Source interface {
	net.Addr
	TryConnect() error
}

// Addr ...
type Addr struct {
	Network string
	IP      net.IP
	Port    int
}

// Service ...
type Service struct {
	ID      string
	Addr    Addr
	ISP     net.IP
	UDP     int
	TCP     int
	ExtData []byte
}

type source struct {
	service Service
	nat     int
}

// Network ...
func (c source) Network() string {
	return c.addr.Network
}

// String ...
func (c source) String() string {
	return net.JoinHostPort(c.addr.IP.String(), strconv.Itoa(c.addr.Port))
}

// TryConnect ...
func (c source) TryConnect() error {
	remote := c.String()
	localPort := LocalPort(c.Network(), c.mappingPort)
	local := LocalAddr(localPort)
	var dial net.Conn
	var err error
	fmt.Println("ping", "local", local, "remote", remote, "network", c.Network(), "mapping", c.mappingPort)
	if c.mappingPort == localPort {
		dial, err = reuse.Dial(c.Network(), local, remote)
	} else {
		if IsUDP(c.Network()) {
			udp, err := net.DialUDP(c.Network(), &net.UDPAddr{}, ParseUDPAddr(remote))
			if err != nil {
				return err
			}
			err = udp.SetDeadline(time.Now().Add(3 * time.Second))
			if err != nil {
				fmt.Println("debug|Ping|SetDeadline", err)
				return err
			}
			_, err = udp.Write([]byte("hello world"))
			if err != nil {
				fmt.Println("debug|Ping|Write", err)
				return err
			}
			data := make([]byte, maxByteSize)
			read, _, err := udp.ReadFromUDP(data)
			if err != nil {
				fmt.Println("debug|Ping|Read", err)
				return err
			}
			fmt.Println("received: ", string(data[:read]))
			return err
		}
		dial, err = net.Dial(c.Network(), remote)
	}

	if err != nil {
		fmt.Println("debug|Ping|Dial", err)
		return err
	}
	_, err = dial.Write([]byte("hello world"))
	if err != nil {
		fmt.Println("debug|Ping|Write", err)
		return err
	}
	data := make([]byte, maxByteSize)
	read, err := dial.Read(data)
	if err != nil {
		fmt.Println("debug|Ping|Read", err)
		return err
	}
	fmt.Println("received: ", string(data[:read]))
	return err
}

// JSON ...
func (addr *Addr) JSON() []byte {
	marshal, err := json.Marshal(addr)
	if err != nil {
		return nil
	}
	return marshal
}

func tryTCP(addr *Addr) error {
	return nil
}

func tryReverse(addr *Addr) error {
	return nil
}

func tryUDP(addr *Addr) error {
	return nil
}
