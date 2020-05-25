package lurker

import (
	"net"
	"time"
)

type tcpConnector struct {
	timeout time.Duration
	conn    net.Conn
	ticker  *time.Ticker
}

func newTCPConnector(conn net.Conn) Connector {
	c := tcpConnector{
		timeout: 5 * time.Second,
	}
	c.conn = conn
	return &c
}

// Reply ...
func (c *tcpConnector) Reply() (err error) {
	close := true
	defer func() {
		if close {
			c.conn.Close()
		}
	}()
	data := make([]byte, maxByteSize)
	if c.timeout != 0 {
		err := c.conn.SetReadDeadline(time.Now().Add(c.timeout))
		if err != nil {
			log.Debugw("debug|Reply|SetReadDeadline", "error", err)
			return err
		}
	}
	n, err := c.conn.Read(data)
	if err != nil {
		log.Debugw("debug|Reply|Read", "error", err)
		return err
	}

	var r HandshakeRequest
	service, err := DecodeHandshakeRequest(data[:n], &r)
	if err != nil {
		log.Debugw("debug|Reply|DecodeHandshakeRequest", "error", err)
		return err
	}
	if !service.KeepConnect {
		go c.KeepConnect()
		close = false
	}

	netAddr := ParseNetAddr(c.conn.RemoteAddr())
	log.Debugw("debug|Reply|ParseNetAddr", "addr", netAddr)
	var resp HandshakeResponse
	resp.Status = HandshakeStatusSuccess
	resp.Data = []byte("Connected")
	if c.timeout != 0 {
		err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
		if err != nil {
			log.Debugw("debug|Reply|SetWriteDeadline", "error", err)
			return err
		}
	}
	_, err = c.conn.Write(resp.JSON())
	if err != nil {
		log.Debugw("debug|Reply|Write", "error", err)
		return err
	}
	return nil
}

// Heartbeat ...
func (c *tcpConnector) KeepConnect() {
	c.ticker = time.NewTicker(time.Second * 30)
	for {
		select {
		case <-c.ticker.C:
			//todo
			return
		}
	}
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

// Run ...
func (c *tcpConnector) Process() {
	var err error
	data := make([]byte, maxByteSize)
	n, err := c.conn.Read(data)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|Read", "error", err)
		return
	}
	handshake, err := ParseHandshake(data[:n])
	if err != nil {
		return
	}
	handshake.Run(c)
	return
}
