package proxy

import (
	"fmt"
	"io"
	"net"
)

const (
	socks5Version = uint8(5)
)

type socks5 struct {
	Authenticate
}

// ListenPort ...
func (s *socks5) ListenPort(port int) (net.Listener, error) {
	tcpAddr := net.TCPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}
	tcpLis, err := net.ListenTCP("tcp", &tcpAddr)
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
   confidentiality, these requests MUST be encapsulated in the method-
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
func (s *socks5) requests(conn net.Conn) (err error) {
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

	return nil
}
