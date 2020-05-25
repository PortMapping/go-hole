package lurker

import "net"

type tcpConnector struct {
	conn net.Conn
}

func newTCPConnector(conn net.Conn) Connector {
	c := tcpConnector{}
	c.conn = conn
	return &c
}
