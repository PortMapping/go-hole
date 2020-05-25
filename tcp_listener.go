package lurker

import (
	"context"
	"fmt"
	"github.com/portmapping/go-reuse"
	"net"

	"github.com/portmapping/lurker/nat"
)

type tcpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	port        int
	mappingPort int
	nat         nat.NAT
	listener    net.Listener
	cfg         *Config
	ready       bool
}

// IsReady ...
func (l *tcpListener) IsReady() bool {
	return l.ready
}

// MappingPort ...
func (l *tcpListener) MappingPort() int {
	return l.mappingPort
}

type tcpHandshake struct {
	conn     net.Conn
	connBack func(s Source)
}

// ConnectCallback ...
func (t *tcpHandshake) ConnectCallback(f func(f Source)) {
	t.connBack = f
}

// Reply ...
func (t *tcpHandshake) Reply() error {
	close := true
	defer func() {
		if close {
			t.conn.Close()
		}
	}()
	data := make([]byte, maxByteSize)
	n, err := t.conn.Read(data)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|Read", "error", err)
		return err
	}
	ip, port := ParseAddr(t.conn.RemoteAddr().String())
	var r HandshakeRequest
	service, err := DecodeHandshakeRequest(data[:n], &r)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|ParseService", "error", err)
		return err
	}
	if !service.KeepConnect {
		close = false
	}

	c := source{
		addr: Addr{
			Protocol: t.conn.RemoteAddr().Network(),
			IP:       ip,
			Port:     port,
		},
		service: service,
	}
	if t.connBack != nil {
		t.connBack(&c)
	}

	netAddr := ParseNetAddr(t.conn.RemoteAddr())
	log.Debugw("debug|getClientFromTCP|ParseNetAddr", "addr", netAddr)
	var resp HandshakeResponse
	resp.Status = HandshakeStatusSuccess
	resp.Data = []byte("Connected")
	_, err = t.conn.Write(resp.JSON())
	if err != nil {
		log.Debugw("debug|getClientFromTCP|write", "error", err)
		return err
	}
	return nil
}

// Pong ...
func (t *tcpHandshake) Pong() error {
	response := HandshakeResponse{
		Status: HandshakeStatusSuccess,
		Data:   []byte("PONG"),
	}
	write, err := t.conn.Write(response.JSON())
	if err != nil {
		return err
	}
	if write == 0 {
		log.Warnw("write pong", "written", 0)
	}
	return nil
}

// Do ...
func (t *tcpHandshake) Do() error {
	data := make([]byte, maxByteSize)
	n, err := t.conn.Read(data)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|Read", "error", err)
		return err
	}
	handshake, err := ParseHandshake(data[:n])
	if err != nil {
		return err
	}
	return handshake.Process(t)
}

// NewTCPListener ...
func NewTCPListener(cfg *Config) Listener {
	tcp := &tcpListener{
		ctx:    nil,
		cancel: nil,
		port:   cfg.TCP,
		cfg:    cfg,
	}
	tcp.ctx, tcp.cancel = context.WithCancel(context.TODO())
	return tcp
}

// Listen ...
func (l *tcpListener) Listen(c chan<- Source) (err error) {
	tcpAddr := LocalTCPAddr(l.port)
	if l.cfg.Secret != nil {
		l.listener, err = reuse.ListenTLS("tcp", DefaultLocalTCPAddr.String(), l.cfg.Secret)
	} else {
		l.listener, err = reuse.ListenTCP("tcp", tcpAddr)
	}
	if err != nil {
		return err
	}
	fmt.Println("listen tcp on address:", tcpAddr.String())
	go listenTCP(l.ctx, l.listener, c)
	if !l.cfg.NAT {
		return nil
	}

	l.nat, err = mapping("tcp", l.cfg.TCP)
	if err != nil {
		return err
	}
	l.ready = true
	return
}

// Stop ...
func (l *tcpListener) Stop() error {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	return nil
}

func listenTCP(ctx context.Context, listener net.Listener, cli chan<- Source) (err error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			acceptTCP, err := listener.Accept()
			if err != nil {
				log.Debugw("debug|getClientFromTCP|Accept", "error", err)
				continue
			}
			go getClientFromTCP(ctx, acceptTCP, cli)
		}
	}
}

func getClientFromTCP(ctx context.Context, conn net.Conn, cli chan<- Source) error {
	close := true
	defer func() {
		if close {
			conn.Close()
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	default:
		t := tcpHandshake{
			conn: conn,
		}
		t.ConnectCallback(func(f Source) {
			cli <- f
		})
		err := t.Do()
		if err != nil {
			return err
		}
		//
		//	ip, port := ParseAddr(conn.RemoteAddr().String())
		//	service, err := DecodeHandshakeRequest(data[:n])
		//	if err != nil {
		//		log.Debugw("debug|getClientFromTCP|ParseService", "error", err)
		//		return err
		//	}
		//	if service.KeepConnect {
		//		close = false
		//	}
		//	c := source{
		//		addr: Addr{
		//			Protocol: conn.RemoteAddr().Network(),
		//			IP:       ip,
		//			Port:     port,
		//		},
		//		service: service,
		//	}
		//	cli <- &c
		//	netAddr := ParseNetAddr(conn.RemoteAddr())
		//
		//	err = tryReverseTCP(&source{addr: *netAddr,
		//		service: Service{
		//			ID:          GlobalID,
		//			KeepConnect: false,
		//		}})
		//	status := 0
		//	if err != nil {
		//		status = -1
		//		log.Debugw("debug|getClientFromTCP|tryReverseTCP", "error", err)
		//	}
		//
		//	r := &ListenResponse{
		//		Status: status,
		//		Addr:   *netAddr,
		//		Error:  err,
		//	}
		//	_, err = conn.Write(r.JSON())
		//	if err != nil {
		//		log.Debugw("debug|getClientFromTCP|write", "error", err)
		//		return err
		//	}
	}
	return nil
}
