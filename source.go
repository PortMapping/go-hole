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

// TryConnect ...
func (s *source) TryConnect() error {
	//remote := s.String()
	//localPort := LocalPort(s.Network(), s.mappingPort)
	//local := LocalAddr(localPort)
	//var dial net.Conn
	//fmt.Println("ping", "local", local, "remote", remote, "network", s.Network(), "mapping", s.mappingPort)
	log.Infow("connect to", "ip", s.addr.String())
	//go func() {
	//	defer wg.Done()
	//	if err := tryUDP(s); err != nil {
	//		log.Errorw("tryUDP|error", "error", err)
	//		return
	//	}
	//}()
	var err error
	if err = tryTCP(s); err == nil {
		return nil
	}
	log.Errorw("tryTCP|error", "error", err)
	//go func() {
	//	defer wg.Done()
	//	if err := tryReverseUDP(s); err != nil {
	//		log.Errorw("tryReverseUDP|error", "error", err)
	//		return
	//	}
	//}()
	if err = tryReverseTCP(s); err == nil {
		return nil
	}
	log.Errorw("tryReverseTCP|error", "error", err)
	if err := tryUDP(s); err == nil {
		return nil
	}
	log.Errorw("tryUDP|error", "error", err)
	if err := tryReverseUDP(s); err != nil {
		return nil
	}
	log.Errorw("tryReverseUDP|error", "error", err)

	return fmt.Errorf("all try connect is failed")

}

func multiPortDialTCP(addr *net.TCPAddr, lports ...int) (net.Conn, error) {
	var lastErr error
	for _, p := range lports {
		tcp, err := reuse.DialTCP("tcp", LocalTCPAddr(p), addr)
		if err != nil {
			lastErr = err
			continue
		}
		return tcp, nil
	}
	return nil, lastErr
}

func tryReverseTCP(s *source) error {
	tcp, err := multiPortDialTCP(s.addr.TCP(), s.service.PortHole, 0)
	if err != nil {
		log.Debugw("debug|tryReverse|DialTCP", "error", err)
		return err
	}
	//never close
	//defer tcp.Close()
	tcp.SetDeadline(time.Now().Add(3 * time.Second))
	s.service.ExtData = []byte("tryReverseTCP")
	s.service.ID = GlobalID
	s.service.KeepConnect = true
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
	tcp, err := multiPortDialTCP(addr.TCP(), s.service.PortHole, 0)
	if err != nil {
		log.Debugw("debug|tryTCP|DialTCP", "error", err)
		return err
	}
	tcp.SetDeadline(time.Now().Add(3 * time.Second))
	defer tcp.Close()
	s.service.ExtData = []byte("tryTCP")
	s.service.ID = GlobalID
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
