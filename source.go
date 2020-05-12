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
	Reverse() error
	Service() Service
	Addr() Addr
}

// Addr ...
type Addr struct {
	Protocol string
	IP       net.IP
	Port     int
}

// Service ...
type Service struct {
	ID          string
	Addr        Addr
	ISP         net.IP
	Local       net.IP
	PortUDP     int
	PortHole    int
	PortTCP     int
	KeepConnect bool
	ExtData     []byte
}

type source struct {
	addr    Addr
	service Service
	support Support
}

// Service ...
func (s source) Service() Service {
	return s.service
}

// Addr ...
func (s source) Addr() Addr {
	return s.addr
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

// IsZero ...
func (addr Addr) IsZero() bool {
	return addr.Protocol == "" && addr.IP.Equal(net.IPv4zero) && addr.Port == 0
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

// NewSource ...
func NewSource(service Service, addr Addr) Source {
	return &source{
		service: service,
		addr:    addr,
	}
}

// String ...
func (s source) String() string {
	return s.addr.String()
}

// Reverse ...
func (s *source) Reverse() error {
	var err error
	if err = tryReverseTCP(s); err == nil {
		return nil
	}
	log.Debugw("tryReverseTCP|error", "error", err)
	if err := tryReverseUDP(s); err != nil {
		return nil
	}
	log.Debugw("tryReverseUDP|error", "error", err)
	return err
}

// TryConnect ...
func (s *source) TryConnect() error {
	log.Infow("connect to", "ip", s.addr.String())
	var err error
	if err = tryTCP(s); err == nil {
		return nil
	}
	log.Debugw("tryTCP|error", "error", err)
	if err = tryReverseTCP(s); err == nil {
		return nil
	}
	log.Debugw("tryReverseTCP|error", "error", err)
	if err := tryUDP(s); err == nil {
		return nil
	}
	log.Debugw("tryUDP|error", "error", err)
	if err := tryReverseUDP(s); err != nil {
		return nil
	}
	log.Debugw("tryReverseUDP|error", "error", err)
	return fmt.Errorf("all try connect is failed")

}

func multiPortDialTCP(addr *net.TCPAddr, timeout time.Duration, lport int) (net.Conn, bool, error) {
	tcp, err := reuse.DialTimeOut("tcp", LocalTCPAddr(lport).String(), addr.String(), timeout)
	if err != nil {
		tcp, err = reuse.DialTimeOut("tcp", LocalTCPAddr(0).String(), addr.String(), timeout)
		if err != nil {
			return nil, false, err
		}
		return tcp, false, nil
	}
	if lport == 0 {
		return tcp, false, nil
	}
	return tcp, true, nil
}

func tryReverseTCP(s *source) error {
	tcp, keep, err := multiPortDialTCP(s.addr.TCP(), 3*time.Second, s.service.PortHole)
	if err != nil {
		log.Debugw("debug|tryReverse|DialTCP", "error", err)
		return err
	}
	//never close
	if !keep {
		defer tcp.Close()
	}
	//tcp.SetDeadline(time.Now().Add(3 * time.Second))
	s.service.ExtData = []byte("tryReverseTCP")
	s.service.ID = GlobalID
	s.service.KeepConnect = keep
	_, err = tcp.Write(s.service.JSON())
	if err != nil {
		log.Debugw("debug|tryReverse|Write", "error", err)
		return err
	}
	data := make([]byte, maxByteSize)
	n, err := tcp.Read(data)
	if err != nil {
		log.Debugw("debug|tryReverse|ReadFromUDP", "error", err)
		return err
	}
	log.Infow("tryReverseTCP received", "address", string(data[:n]))
	return nil
}

func tryReverseUDP(s *source) error {
	udp, err := net.DialUDP("udp", LocalUDPAddr(s.service.PortHole), s.addr.UDP())
	if err != nil {
		log.Debugw("debug|tryReverse|DialUDP", "error", err)
		return err
	}
	s.service.ExtData = []byte("tryReverseUDP")
	_, err = udp.Write(s.service.JSON())
	if err != nil {
		log.Debugw("debug|tryReverse|Write", "error", err)
		return err
	}
	data := make([]byte, maxByteSize)
	n, _, err := udp.ReadFromUDP(data)
	if err != nil {
		log.Debugw("debug|tryReverse|ReadFromUDP", "error", err)
		return err
	}
	log.Infow("tryReverseUDP received", "address", string(data[:n]))
	return err
}

func tryUDP(s *source) error {
	udp, err := net.DialUDP("udp", LocalUDPAddr(s.service.PortHole), s.addr.UDP())
	if err != nil {
		log.Debugw("debug|tryUDP|DialUDP", "error", err)
		return err
	}
	s.service.ExtData = []byte("tryUDP")
	_, err = udp.Write(s.service.JSON())
	if err != nil {
		log.Debugw("debug|tryUDP|Write", "error", err)
		return err
	}
	data := make([]byte, maxByteSize)
	n, remote, err := udp.ReadFromUDP(data)
	if err != nil {
		log.Debugw("debug|tryUDP|ReadFromUDP", "error", err)
		return err
	}
	log.Infow("tryUDP received", "remote info", remote.String(), "address", string(data[:n]))
	return nil
}

func tryTCP(s *source) error {
	addr := ParseSourceAddr("tcp", s.addr.IP, s.service.PortTCP)
	//tcp, err := net.Dial("tcp", tcpAddr.String())
	tcp, keep, err := multiPortDialTCP(addr.TCP(), 3*time.Second, s.service.PortHole)
	if err != nil {
		log.Debugw("debug|tryTCP|DialTCP", "error", err)
		return err
	}
	if !keep {
		defer tcp.Close()
	}
	s.service.ExtData = []byte("tryTCP")
	s.service.ID = GlobalID
	s.service.KeepConnect = true
	_, err = tcp.Write(s.service.JSON())
	if err != nil {
		log.Debugw("debug|tryTCP|Write", "error", err)
		return err
	}
	data := make([]byte, maxByteSize)
	n, err := tcp.Read(data)
	if err != nil {
		log.Debugw("debug|tryTCP|ReadFromUDP", "error", err)
		return err
	}
	log.Infow("tryTCP received", "address", string(data[:n]))

	return nil
}

// ParseSourceAddr ...
func ParseSourceAddr(network string, ip net.IP, port int) *Addr {
	return &Addr{
		Protocol: network,
		IP:       ip,
		Port:     port,
	}
}
