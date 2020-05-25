package lurker

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/portmapping/go-reuse"
)

// Source ...
type Source interface {
	Connect() error
	Try() error
	Service() Service
	Addr() Addr
	SetMappingPort(string, int) //T.B.D
}

type source struct {
	mappingPortUDP int
	mappingPortTCP int
	service        Service
	addr           Addr
	support        Support
	timeout        time.Duration
}

// SetMappingPort ...
func (s *source) SetMappingPort(network string, i int) {
	switch network {
	case "tcp", "tcp6", "tcp4":
		s.mappingPortTCP = i
	case "udp", "udp6", "udp4":
		s.mappingPortUDP = i
	}
}

// service ...
func (s source) Service() Service {
	return s.service
}

// Addr ...
func (s source) Addr() Addr {
	return s.addr
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

// Try ...
func (s *source) Try() error {
	log.Infow("connect to", "ip", s.addr.String())
	defer func() {
		fmt.Println("supported", s.support.List)
	}()
	var err error
	//var addr *Addr
	if err = tryPublicNetworkConnect(s); err != nil {
		log.Debugw("debug|tryPublicNetworkConnect|error")
	}

	if err := tryReverseNetworkConnect(s); err != nil {
		log.Debugw("debug|tryReverseNetworkConnect|error", "error", err)
	}

	return fmt.Errorf("all try connect is failed")

}

// Connect ...
func (s *source) Connect() error {
	log.Infow("connect to", "ip", s.addr.String())

	var err error
	err = tryConnect(s, &s.addr)
	if err != nil {
		return err
	}
	return nil
}

func tryReverseNetworkConnect(s *source) error {
	switch s.addr.Network() {
	case "tcp", "tcp4", "tcp6":
		tcpAddr := ParseSourceAddr(s.addr.Protocol, s.addr.IP, s.addr.Port)
		if err := tryTCP(s, tcpAddr); err != nil {
			return err
		}
		s.support.List[ProviderNetworkTCP] = true
	case "udp", "udp4", "udp6":
		udpAddr := ParseSourceAddr(s.addr.Protocol, s.addr.IP, s.addr.Port)
		if err := tryUDP(s, udpAddr); err != nil {
			return err
		}
		s.support.List[ProviderNetworkUDP] = true
	default:
		return fmt.Errorf("no reverse service found")
	}
	return nil
}

func tryConnect(s *source, addr *Addr) error {
	switch s.addr.Network() {
	case "tcp", "tcp4", "tcp6":
		//tcpAddr := ParseSourceAddr(addr.Protocol, addr.IP, addr.Port)
		tcpAddr, _, err := multiPortDialTCP(addr.TCP(), s.timeout, s.mappingPortTCP)
		if err != nil {
			log.Debugw("debug|tryUDP|DialUDP", "error", err)
			return err
		}
		data := make([]byte, maxByteSize)

		if _, err := tcpConnect(s, tcpAddr, data); err != nil {
			return err
		}
		s.support.List[ProviderNetworkTCP] = true
	case "udp", "udp4", "udp6":
		udpAddr := ParseSourceAddr(addr.Protocol, addr.IP, addr.Port)

		if err := tryUDP(s, udpAddr); err != nil {
			return err
		}
		s.support.List[ProviderNetworkUDP] = true
	default:
		return fmt.Errorf("no reverse service found")
	}
	return nil
}

func tryPublicNetworkConnect(s *source) error {
	//switch s.addr.Network() {
	//case "tcp", "tcp4", "tcp6":
	tcpAddr := ParseSourceAddr("tcp", s.addr.IP, s.service.PortTCP)
	if err := tryTCP(s, tcpAddr); err != nil {
		log.Debugw("debug|tryPublicNetworkConnect|tryTCP", "error", err)
	} else {
		s.support.List[PublicNetworkTCP] = true
	}

	//case "udp", "udp4", "udp6":
	udpAddr := ParseSourceAddr("udp", s.addr.IP, s.service.PortUDP)
	if err := tryUDP(s, udpAddr); err != nil {
		log.Debugw("debug|tryPublicNetworkConnect|tryUDP", "error", err)
	} else {
		s.support.List[PublicNetworkUDP] = true
	}
	log.Debugw("tryPublicNetworkConnect|finished")
	return nil
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
	tcp, keep, err := multiPortDialTCP(s.addr.TCP(), 3*time.Second, s.mappingPortTCP)
	if err != nil {
		log.Debugw("debug|tryReverseNetworkConnect|DialTCP", "error", err)
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
	n, err := tcpPing(s, tcp, data)
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
	udp, err := multiPortDialUDP(s.addr.UDP(), s.mappingPortUDP)
	if err != nil {
		log.Debugw("debug|tryReverseUDP|DialUDP", "error", err)
		return err
	}
	//s.service.ExtData = []byte("tryReverseUDP")
	data := make([]byte, maxByteSize)
	n, err := udpPing(s, udp, data)
	if err != nil {
		return err
	}
	//ignore n
	_ = n
	return nil
}

func tryUDP(s *source, addr *Addr) error {
	udp, err := multiPortDialUDP(addr.UDP(), s.mappingPortUDP)
	if err != nil {
		log.Debugw("debug|tryUDP|DialUDP", "error", err)
		return err
	}
	data := make([]byte, maxByteSize)
	n, err := udpPing(s, udp, data)
	if err != nil {
		return err
	}
	//ignore n
	_ = n
	return nil
}
func tcpConnect(s *source, conn net.Conn, data []byte) (n int, err error) {
	if s.timeout != 0 {
		err = conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	handshake := HandshakeHead{
		Type: HandshakeTypeConnect,
	}
	_, err = conn.Write(handshake.JSON())
	if err != nil {
		log.Debugw("debug|tcpConnect|Write", "error", err)
		return 0, err
	}
	if s.timeout != 0 {
		err = conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	req, err := EncodeHandshakeRequest(s.service)
	if err != nil {
		return 0, err
	}
	_, err = conn.Write(req)
	if err != nil {
		log.Debugw("debug|tcpConnect|Write", "error", err)
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
		log.Debugw("debug|tcpConnect|Read", "error", err)
		return 0, err
	}
	log.Infow("tcp received", "data", string(data[:n]))
	return n, nil
}

func tcpPing(s *source, conn net.Conn, data []byte) (n int, err error) {
	if s.timeout != 0 {
		err = conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	handshake := HandshakeHead{
		Type: HandshakeTypePing,
	}

	_, err = conn.Write(handshake.JSON())
	if err != nil {
		log.Debugw("debug|tcpPing|Write", "error", err)
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
		log.Debugw("debug|tcpPing|ReadFromUDP", "error", err)
		return 0, err
	}
	log.Infow("tcp received", "data", string(data[:n]))
	return n, nil
}
func udpPing(s *source, conn *net.UDPConn, data []byte) (n int, err error) {
	if s.timeout != 0 {
		err = conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	_, err = conn.Write(s.service.JSON())
	if err != nil {
		log.Debugw("debug|udpPing|Write", "error", err)
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
		log.Debugw("debug|udpPing|ReadFromUDP", "error", errR)
		return 0, errR
	}
	log.Infow("udp received", "remote info", remote.String(), "data", string(data[:n]))
	return n, nil
}

func tryTCP(s *source, addr *Addr) error {
	log.Debugw("connect tcp", "port", s.mappingPortTCP, "addr", addr.TCP().String())
	tcp, keep, err := multiPortDialTCP(addr.TCP(), 5*time.Second, s.mappingPortTCP)
	if err != nil {
		log.Debugw("debug|tryTCP|DialTCP", "error", err)
		return err
	}
	if !keep {
		defer tcp.Close()
	}
	s.service.ID = GlobalID
	s.service.KeepConnect = true
	data := make([]byte, maxByteSize)
	n, err := tcpPing(s, tcp, data)
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
