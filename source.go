package lurker

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/portmapping/go-reuse"
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

type source struct {
	service Service
	addr    Addr
	support Support
	timeout time.Duration
}

// service ...
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

// NewSource ...
func NewSource(service Service, addr Addr) Source {
	return &source{
		service: service,
		addr:    addr,
		timeout: DefaultConnectionTimeout,
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
	if err = tryPublicNetworkConnect(s); err == nil {
		log.Debugw("tryPublicNetworkConnect|success")
		return nil
	}

	log.Debugw("tryPublicNetworkConnect|error", "error", err)
	if err = tryPublicNetworkTCP(s); err == nil {
		log.Debugw("tryPublicNetworkTCP|success")
		return nil
	}
	log.Debugw("tryPublicNetworkTCP|error", "error", err)
	if err = tryReverseTCP(s); err == nil {
		log.Debugw("tryReverseTCP|success")
		return nil
	}
	log.Debugw("tryReverseTCP|error", "error", err)
	if err := tryPublicNetworkUDP(s); err == nil {
		log.Debugw("tryPublicNetworkUDP|success")
		return nil
	}
	log.Debugw("tryPublicNetworkUDP|error", "error", err)
	if err := tryReverseUDP(s); err != nil {
		log.Debugw("tryReverseUDP|success")
		return nil
	}
	log.Debugw("tryReverseUDP|error", "error", err)
	return fmt.Errorf("all try connect is failed")

}

func tryPublicNetworkConnect(s *source) error {
	switch s.addr.Network() {
	case "tcp", "tcp4", "tcp6":
		addr := ParseSourceAddr("tcp", s.addr.IP, s.service.PortTCP)
		return tryPublicNetworkTCP(s, addr)
	case "udp", "udp4", "udp6":
		return tryPublicNetworkUDP(s)
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

	//s.service.ExtData = []byte("tryReverseTCP")
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

func multiPortDialUDP(addr *net.UDPAddr, lport int) (*net.UDPConn, error) {
	udp, err := net.DialUDP("udp", LocalUDPAddr(lport), addr)
	if err != nil {
		udp, err = net.DialUDP("udp", LocalUDPAddr(0), addr)
		if err != nil {
			return nil, err
		}
	}
	return udp, nil
}

func tryReverseUDP(s *source) error {
	udp, err := multiPortDialUDP(s.addr.UDP(), s.service.PortHole)
	if err != nil {
		log.Debugw("debug|tryReverseUDP|DialUDP", "error", err)
		return err
	}
	//s.service.ExtData = []byte("tryReverseUDP")
	data := make([]byte, maxByteSize)
	n, err := udpRW(s, udp, data)
	if err != nil {
		return err
	}
	//ignore n
	_ = n
	return nil
}

func tryPublicNetworkUDP(s *source) error {
	addr := ParseSourceAddr("udp", s.addr.IP, s.service.PortUDP)
	udp, err := multiPortDialUDP(addr.UDP(), s.service.PortHole)
	if err != nil {
		log.Debugw("debug|tryPublicNetworkUDP|DialUDP", "error", err)
		return err
	}
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
	handshake := Handshake{
		Type: HandshakeTypePing,
	}
	_, err = conn.Write(handshake.JSON())
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
	log.Infow("tcp received", "data", string(data[:n]))
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
	n, remote, errR := conn.ReadFromUDP(data)
	if errR != nil {
		log.Debugw("debug|udpRW|ReadFromUDP", "error", errR)
		return 0, errR
	}
	log.Infow("udp received", "remote info", remote.String(), "data", string(data[:n]))
	return n, nil
}

func tryPublicNetworkTCP(s *source, addr Addr) error {
	tcp, keep, err := multiPortDialTCP(addr.TCP(), 3*time.Second, s.service.PortHole)
	if err != nil {
		log.Debugw("debug|tryPublicNetworkTCP|DialTCP", "error", err)
		return err
	}
	if !keep {
		defer tcp.Close()
	}
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
