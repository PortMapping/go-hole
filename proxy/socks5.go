package proxy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/portmapping/go-reuse"
	"io"
	"net"
	"strconv"
)

const (
	socks5Version = uint8(5)
	rsvRESERVED   = 0x00
)
const (
	cmdConnect      = 0x01
	cmdBind         = 0x02
	cmdUDPAssociate = 0x03
)
const (
	repSucceeded = iota
	repGeneralSOCKSServerFailure
	repConnectionNotAlloweByRuleset
	repNetworkUnreachable
	repHostUnreachable
	repConnectionRefused
	repTTLExpired
	repCommandNotSupported
	repAddressTypeNotSupported
	repUnassigned
	repEnd = 0xFF
)

const (
	atypIPv4Address = 0x01
	atypDomainName  = 0x03
	atypIPv6Address = 0x04
)

type socks5 struct {
	Authenticate
}

var errAddressTypeNotSupported = errors.New("address type not supported")

// ListenPort ...
func (s *socks5) ListenPort(port int) (net.Listener, error) {
	tcpAddr := net.TCPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}
	tcpLis, err := reuse.ListenTCP("tcp", &tcpAddr)
	if err != nil {
		return nil, err
	}
	return tcpLis, nil
}

// ListenTCP ...
func (s *socks5) Monitor(conn net.Conn) {
	if err := s.listenRequest(conn); err != nil {
		return
	}
	if err := s.doRequests(conn); err != nil {
		return
	}
}

func newSocks5Proxy(auth Authenticate) (Proxy, error) {
	return &socks5{
		Authenticate: auth,
	}, nil
}

// ListenRequest ...
func (s *socks5) listenRequest(conn net.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("failed read socks5 data: %w", err)
	}

	if version := buf[0]; version != socks5Version {
		return fmt.Errorf("wrong socks5 version: %v", version)
	}
	nMethods := buf[1]
	methods := make([]byte, nMethods)
	if len, err := conn.Read(methods); len != int(nMethods) || err != nil {
		return fmt.Errorf("wrong nmethod: %w", err)
	}
	if s.NeedAuthenticate() {
		buf[1] = userPassAuth
		conn.Write(buf)
		if err := s.Auth(conn); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	} else {
		buf[1] = 0
		conn.Write(buf)
	}
	return nil
}

/*
Requests
   Once the method-dependent subnegotiation has completed, the client
   sends the request details.  If the negotiated method includes
   encapsulation for purposes of integrity checking and/or
   confidentiality, these doRequests MUST be encapsulated in the method-
   dependent encapsulation.

   The SOCKS request is formed as follows:

        +----+-----+-------+------+----------+----------+
        |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
        +----+-----+-------+------+----------+----------+
        | 1  |  1  | X'00' |  1   | Variable |    2     |
        +----+-----+-------+------+----------+----------+

     Where:

          o  VER    protocol version: X'05'
          o  CMD
             o  CONNECT X'01'
             o  BIND X'02'
             o  UDP ASSOCIATE X'03'
          o  RSV    RESERVED
          o  ATYP   address type of following address
             o  IP V4 address: X'01'
             o  DOMAINNAME: X'03'
             o  IP V6 address: X'04'
          o  DST.ADDR       desired destination address
          o  DST.PORT desired destination port in network octet
             order

   The SOCKS server will typically evaluate the request based on source
   and destination addresses, and return one or more reply messages, as
   appropriate for the request type.
*/
func (s *socks5) doRequests(conn net.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	header := make([]byte, 3)

	_, err = io.ReadFull(conn, header)
	if err != nil {
		return
	}
	fmt.Println("receive new data")
	switch header[1] {
	case cmdConnect:
		_, e := getAddrPort(conn)
		if e != nil {
			err = doReplies(conn, repAddressTypeNotSupported, atypIPv4Address)
			return
		}

	case cmdBind:
	case cmdUDPAssociate:
	}
	conn.Close()
	return nil
}

func getAddrPort(conn net.Conn) (addr string, err error) {
	addrType := make([]byte, 1)
	conn.Read(addrType)
	var host string
	switch addrType[0] {
	case atypIPv4Address:
		ipv4 := make(net.IP, net.IPv4len)
		conn.Read(ipv4)
		host = ipv4.String()
	case atypIPv6Address:
		ipv6 := make(net.IP, net.IPv6len)
		conn.Read(ipv6)
		host = ipv6.String()
	case atypDomainName:
		var domainLen uint8
		binary.Read(conn, binary.BigEndian, &domainLen)
		domain := make([]byte, domainLen)
		conn.Read(domain)
		host = string(domain)
	default:
		return "", errAddressTypeNotSupported
	}

	var port uint16
	binary.Read(conn, binary.BigEndian, &port)
	// connect to host
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

func connect(cmd int, conn net.Conn) {

}

func doReplies(conn net.Conn, rep byte, atyp byte) (err error) {
	reply := []byte{
		socks5Version,
		rep,
		rsvRESERVED,
		atyp,
	}

	localAddr := conn.LocalAddr().String()
	localHost, localPort, _ := net.SplitHostPort(localAddr)
	ipBytes := net.ParseIP(localHost).To4()
	nPort, _ := strconv.Atoi(localPort)
	reply = append(reply, ipBytes...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(nPort))
	reply = append(reply, portBytes...)

	_, err = conn.Write(reply)
	return
}
