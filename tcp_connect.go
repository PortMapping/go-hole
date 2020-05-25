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

// Reply ...
func (c *tcpConnector) Reply() (err error) {
	close := true
	defer func() {
		if close {
			c.conn.Close()
		} else {
			go c.Heartbeat()
		}
	}()
	data := make([]byte, maxByteSize)
	n, err := c.conn.Read(data)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|Read", "error", err)
		return err
	}

	var r HandshakeRequest
	service, err := DecodeHandshakeRequest(data[:n], &r)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|ParseService", "error", err)
		return err
	}
	if !service.KeepConnect {
		close = false
	}

	netAddr := ParseNetAddr(c.conn.RemoteAddr())
	log.Debugw("debug|getClientFromTCP|ParseNetAddr", "addr", netAddr)
	var resp HandshakeResponse
	resp.Status = HandshakeStatusSuccess
	resp.Data = []byte("Connected")
	_, err = c.conn.Write(resp.JSON())
	if err != nil {
		log.Debugw("debug|getClientFromTCP|write", "error", err)
		return err
	}
	return nil
}

// Heartbeat ...
func (c *tcpConnector) Heartbeat() {

}

// Pong ...
func (c *tcpConnector) Pong() error {
	defer c.conn.Close()
	response := HandshakeResponse{
		Status: HandshakeStatusSuccess,
		Data:   []byte("PONG"),
	}
	write, err := c.conn.Write(response.JSON())
	if err != nil {
		return err
	}
	if write == 0 {
		log.Warnw("write pong", "written", 0)
	}
	return nil
}

// Process ...
func (c *tcpConnector) Process() error {
	return nil
}
