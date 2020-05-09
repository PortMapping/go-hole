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
	TryConnect() error
}

// Addr ...
type Addr struct {
	Protocol string
	IP       net.IP
	Port     int
}

// Service ...
type Service struct {
	ID       string
	ISP      net.IP
	UDP      int
	HolePort int
	TCP      int
	ExtData  []byte
}

type source struct {
	addr    Addr
	service Service
	nat     int
}

// Network ...
func (addr Addr) Network() string {
	return addr.Protocol
}

// Network ...
func (addr Addr) String() string {
	return net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port))
}

// UDP ...
func (addr Addr) UDP() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   addr.IP,
		Port: addr.Port,
	}
}

// TCP ...
func (addr Addr) TCP() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   addr.IP,
		Port: addr.Port,
	}
}

// JSON ...
func (s Service) JSON() []byte {
	marshal, err := json.Marshal(s)
	if err != nil {
		return nil
	}
	return marshal
}

// ParseService ...
func ParseService(data []byte) (service Service, err error) {
	err = json.Unmarshal(data, &service)
	return
}

// String ...
func (c source) String() string {
	return c.addr.String()
}

// TryConnect ...
func (c source) TryConnect() error {
	remote := c.String()
	//localPort := LocalPort(c.Network(), c.mappingPort)
	//local := LocalAddr(localPort)
	var dial net.Conn
	var err error
	//fmt.Println("ping", "local", local, "remote", remote, "network", c.Network(), "mapping", c.mappingPort)
	err = tryReverse(&c)
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

func tryTCP(addr *Addr) error {
	return nil
}

func tryReverse(s *source) error {
	udp, err := net.DialUDP("udp", LocalUDPAddr(s.service.HolePort), s.addr.UDP())
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

func tryUDP(addr *Addr) error {
	return nil
}

// ParseSourceAddr ...
func ParseSourceAddr(network string, ip net.IP, port int) *Addr {
	net.TCPAddr{
		IP:   nil,
		Port: 0,
		Zone: "",
	}
	return &Addr{
		Network: network,
		IP:      ip,
		Port:    port,
	}
}
