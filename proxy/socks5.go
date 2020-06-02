package proxy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/goextension/log"
	"github.com/panjf2000/ants/v2"
	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/common"
	"github.com/portmapping/lurker/nat"
	"github.com/portmapping/lurker/pool"
	"io"
	"net"
	"strconv"
	"sync"
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
	nat      nat.NAT
	funcPool *ants.PoolWithFunc
}

var errAddressTypeNotSupported = errors.New("common type not supported")
var errCommandNotSupported = errors.New("command not supported")

// ListenOnPort ...
func (s *socks5) ListenOnPort(port int) (net.Listener, error) {
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

// Connect ...
func (s *socks5) Connect(conn net.Conn) error {
	return s.funcPool.Invoke(conn)
}

func newSocks5Proxy(n nat.NAT, auth Authenticate) (Proxy, error) {

	s := &socks5{
		nat:          n,
		Authenticate: auth,
	}
	funcPool, err := ants.NewPoolWithFunc(ants.DefaultAntsPoolSize, s.handleConnect, ants.WithNonblocking(false))
	if err != nil {
		return nil, err
	}
	s.funcPool = funcPool
	return s, nil
}

func (s *socks5) procedureProc(conn net.Conn) (err error) {
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
          o  ATYP   common type of following common
             o  IP V4 common: X'01'
             o  DOMAINNAME: X'03'
             o  IP V6 common: X'04'
          o  DST.ADDR       desired destination common
          o  DST.PORT desired destination port in network octet
             order

   The SOCKS server will typically evaluate the request based on source
   and destination addresses, and return one or more reply messages, as
   appropriate for the request type.
*/
func doRequests(conn net.Conn) (err error) {
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
	switch header[1] {
	case cmdConnect:
		log.Debugw("proxy connect")
		e := connect(cmdConnect, conn)
		if e == errAddressTypeNotSupported {
			err = doReplies(conn, repAddressTypeNotSupported, atypIPv4Address)
			if err != nil {
				fmt.Println("error", err)
			}
		}
	case cmdBind:
	case cmdUDPAssociate:
	}
	conn.Write([]byte{socks5Version, repCommandNotSupported})
	conn.Close()
	return nil
}

func (s *socks5) handleConnect(i interface{}) {
	fmt.Println("handle Connect")
	conn, b := i.(net.Conn)
	if !b {
		return
	}
	if err := s.procedureProc(conn); err != nil {
		return
	}
	if err := doRequests(conn); err != nil {
		return
	}
}

func getAddrPort(conn net.Conn) (addr string, err error) {
	addrType := make([]byte, 1)
	_, err = conn.Read(addrType)
	if err != nil {
		return "", err
	}
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
	log.Debugw("proxy get addr")
	var port uint16
	err = binary.Read(conn, binary.BigEndian, &port)
	if err != nil {
		return "", err
	}
	// connect to host
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

func connect(cmd int, conn net.Conn) error {
	addr, e := getAddrPort(conn)
	if e != nil {
		return e
	}
	tcpAddr, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		return e
	}
	localTCPAddr := common.LocalTCPAddr(0)
	dial, err := net.DialTCP("tcp", localTCPAddr, tcpAddr)
	if err != nil {
		return err
	}

	e = doReplies(conn, repSucceeded, atypIPv4Address)
	if e != nil {
		return e
	}
	wg := sync.WaitGroup{}

	wg.Add(1)
	pool.AddConnections(pool.NewConnection(conn, dial, &wg))
	wg.Wait()
	return nil
}

func doReplies(conn net.Conn, rep byte, atyp byte) (err error) {
	reply := []byte{
		socks5Version,
		rep,
		rsvRESERVED,
		atyp,
	}
	addr := common.ParseNetAddr(conn.LocalAddr())
	reply = append(reply, addr.IP.To4()...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(addr.Port))
	reply = append(reply, portBytes...)

	_, err = conn.Write(reply)
	return
}
