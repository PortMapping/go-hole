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
	Protocol string `json:"protocol"`
	IP       net.IP `json:"ip"`
	Port     int    `json:"port"`
}

// Service ...
type Service struct {
	ID          string `json:"id"`
	Addr        []Addr `json:"addr"`
	ISP         net.IP `json:"isp"`
	Local       net.IP `json:"local"`
	PortUDP     int    `json:"port_udp"`
	PortHole    int    `json:"port_hole"`
	PortTCP     int    `json:"port_tcp"`
	KeepConnect bool   `json:"keep_connect"`
	ExtData     []byte `json:"ext_data"`
}

type source struct {
	addr    Addr
	service Service
	support Support
	timeout time.Duration
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
		timeout: time.Duration(DefaultTimeout),
	}
}

// String ...
func (s source) String() string {
	return s.addr.String()
}

// TryConnect ...
func (s *source) TryConnect() error {
	log.Infow("connect to", "ip", s.addr.String())
	var err error
	if err = tryConnect(s); err == nil {
		log.Debugw("tryConnect|success")
		return nil
	}
	log.Debugw("tryConnect|error", "error", err)
	if err = tryTCP(s); err == nil {
		log.Debugw("tryTCP|success")
		return nil
	}
	log.Debugw("tryTCP|error", "error", err)
	if err = tryReverseTCP(s); err == nil {
		log.Debugw("tryReverseTCP|success")
		return nil
	}
	log.Debugw("tryReverseTCP|error", "error", err)
	if err := tryUDP(s); err == nil {
		log.Debugw("tryUDP|success")
		return nil
	}
	log.Debugw("tryUDP|error", "error", err)
	if err := tryReverseUDP(s); err != nil {
		log.Debugw("tryReverseUDP|success")
		return nil
	}
	log.Debugw("tryReverseUDP|error", "error", err)
	return fmt.Errorf("all try connect is failed")

}

func tryConnect(s *source) error {
	switch s.addr.Network() {
	case "tcp":
		return tryTCP(s)
	case "udp":
		return tryUDP(s)
	default:
	}
	return fmt.Errorf("network not supported")
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

	s.service.ExtData = []byte("tryReverseTCP")
	s.service.ID = GlobalID
	s.service.KeepConnect = keep
	data := make([]byte, maxByteSize)
	n, err := tcpRW(s, tcp, data)
	if err != nil {
		return err
	}
	//ignore n
	_ = n
	return nil
}

func tryReverseUDP(s *source) error {
	udp, err := net.DialUDP("udp", LocalUDPAddr(s.service.PortHole), s.addr.UDP())
	if err != nil {
		log.Debugw("debug|tryReverse|DialUDP", "error", err)
		return err
	}
	s.service.ExtData = []byte("tryReverseUDP")
	data := make([]byte, maxByteSize)
	n, err := udpRW(s, udp, data)
	if err != nil {
		return err
	}
	//ignore n
	_ = n
	return err
}

func tryUDP(s *source) error {
	udp, err := net.DialUDP("udp", LocalUDPAddr(s.service.PortHole), s.addr.UDP())
	if err != nil {
		log.Debugw("debug|tryUDP|DialUDP", "error", err)
		return err
	}
	s.service.ExtData = []byte("tryUDP")
	data := make([]byte, maxByteSize)
	n, err := udpRW(s, udp, data)
	if err != nil {
		return err
	}
	//ignore n
	_ = n
	return nil
}
func tcpRW(s *source, conn net.Conn, data []byte) (n int, err error) {
	if s.timeout != 0 {
		err = conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	_, err = conn.Write(s.service.JSON())
	if err != nil {
		log.Debugw("debug|tcpRW|Write", "error", err)
		return 0, err
	}
	if s.timeout != 0 {
		err = conn.SetReadDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	n, err = conn.Read(data)
	if err != nil {
		log.Debugw("debug|tcpRW|ReadFromUDP", "error", err)
		return 0, err
	}
	log.Infow("udp received", "data", string(data[:n]))
	return n, nil
}
func udpRW(s *source, conn *net.UDPConn, data []byte) (n int, err error) {
	if s.timeout != 0 {
		err = conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	_, err = conn.Write(s.service.JSON())
	if err != nil {
		log.Debugw("debug|udpRW|Write", "error", err)
		return 0, err
	}
	//data := make([]byte, maxByteSize)
	if s.timeout != 0 {
		err = conn.SetReadDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	n, remote, err := conn.ReadFromUDP(data)
	if err != nil {
		log.Debugw("debug|udpRW|ReadFromUDP", "error", err)
		return 0, err
	}
	log.Infow("udp received", "remote info", remote.String(), "data", string(data[:n]))
	return n, nil
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
	data := make([]byte, maxByteSize)
	n, err := tcpRW(s, tcp, data)
	if err != nil {
		return err
	}
	//ignore n
	_ = n
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

// ParseNetAddr ...
func ParseNetAddr(addr net.Addr) *Addr {
	ip, port := ParseAddr(addr.String())
	return &Addr{
		Protocol: addr.Network(),
		IP:       ip,
		Port:     port,
	}
}
