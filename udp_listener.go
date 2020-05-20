package lurker

import (
	"context"
	"fmt"
	p2pnat "github.com/libp2p/go-nat"
	"github.com/portmapping/lurker/nat"
	"net"
)

type udpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	port        int
	mappingPort int
	nat         nat.NAT
	udpListener *net.UDPConn
	cfg         *Config
}

// Stop ...
func (l *udpListener) Stop() error {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	return nil
}

// NewUDPListener ...
func NewUDPListener(cfg *Config) Listener {
	udp := &udpListener{
		ctx:    nil,
		cancel: nil,
		cfg:    cfg,
		port:   cfg.UDP,
	}
	udp.ctx, udp.cancel = context.WithCancel(context.TODO())
	return udp
}

// Listen ...
func (l *udpListener) Listen(c chan<- Source) (err error) {
	udpAddr := LocalUDPAddr(l.port)
	l.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	fmt.Println("listen udp on address:", udpAddr.String())
	go listenUDP(l.ctx, l.udpListener, c)

	if !l.cfg.NAT {
		return nil
	}

	l.nat, err = nat.FromLocal("udp", l.cfg.UDP)
	if err != nil {
		log.Debugw("nat error", "error", err)
		if err == p2pnat.ErrNoNATFound {

		}
		l.cfg.NAT = false
	} else {
		extPort, err := l.nat.Mapping()
		if err != nil {
			log.Debugw("nat mapping error", "error", err)
			l.cfg.NAT = false
			return nil
		}
		l.mappingPort = extPort

		address, err := l.nat.GetExternalAddress()
		if err != nil {
			log.Debugw("get external address error", "error", err)
			l.cfg.NAT = false
			return nil
		}
		addr := ParseSourceAddr("tcp", address, extPort)
		fmt.Println("udp mapping on address:", addr.String())
	}
	return nil
}

type udpHandshake struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	connBack func(f Source)
}

// Pong ...
func (h *udpHandshake) Pong() error {
	response := HandshakeResponse{
		Status: HandshakeStatusSuccess,
		Data:   []byte("PONG"),
	}
	write, err := h.conn.WriteToUDP(response.JSON(), h.addr)
	if err != nil {
		return err
	}
	if write == 0 {
		log.Warnw("write pong", "written", 0)
	}
	return nil
}

// Reply ...
func (h *udpHandshake) Reply() error {
	data := make([]byte, maxByteSize)
	n, err := h.conn.Read(data)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|Read", "error", err)
		return err
	}
	ip, port := ParseAddr(h.conn.RemoteAddr().String())
	var r HandshakeRequest
	service, err := DecodeHandshakeRequest(data[:n], &r)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|ParseService", "error", err)
		return err
	}

	c := source{
		addr: Addr{
			Protocol: h.addr.Network(),
			IP:       ip,
			Port:     port,
		},
		service: service,
	}
	h.connBack(&c)

	netAddr := ParseNetAddr(h.addr)
	log.Debugw("debug|getClientFromTCP|ParseNetAddr", netAddr)
	var resp HandshakeResponse
	resp.Status = HandshakeStatusSuccess
	resp.Data = []byte("Connected")
	_, err = h.conn.WriteToUDP(resp.JSON(), h.addr)
	if err != nil {
		log.Debugw("debug|getClientFromTCP|write", "error", err)
		return err
	}
	return nil
}

// ConnectCallback ...
func (h *udpHandshake) ConnectCallback(f func(f Source)) {
	h.connBack = f
}

// Do ...
func (h *udpHandshake) Do() (err error) {
	data := make([]byte, maxByteSize)
	var n int
	n, h.addr, err = h.conn.ReadFromUDP(data)
	if err != nil {
		//waiting for next
		log.Debugw("debug|listenUDP|ReadFromUDP", "error", err)
		return err
	}
	handshake, err := ParseHandshake(data[:n])
	if err != nil {
		return err
	}
	return handshake.Process(h)
}

func listenUDP(ctx context.Context, listener *net.UDPConn, cli chan<- Source) (err error) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			u := udpHandshake{
				conn: listener,
			}
			u.ConnectCallback(func(f Source) {
				cli <- f
			})
			err = u.Do()
			if err != nil {
				continue
			}

			return nil
			//handshake, err := ParseHandshake(data)
			//if err != nil {
			//	return err
			//}
			//var req HandshakeRequest
			//service, err := DecodeHandshakeRequest(data[:n], &req)
			//if err != nil {
			//	//waiting for next
			//	log.Debugw("debug|listenUDP|ParseService", "error", err)
			//	continue
			//}

			//netAddr := ParseNetAddr(remoteAddr)
			//c := source{
			//	addr:    *netAddr,
			//	service: service,
			//}
			//cli <- &c
			//err = tryReverseUDP(&source{
			//	addr: *netAddr,
			//	service: Service{
			//		ID:          GlobalID,
			//		KeepConnect: false,
			//	}})
			//status := 0
			//if err != nil {
			//	status = -1
			//	log.Debugw("debug|listenUDP|tryReverseUDP", "error", err)
			//}
			//
			//r := &ListenResponse{
			//	Status: status,
			//	Addr:   *netAddr,
			//	Error:  err,
			//}
			//_, err = listener.WriteToUDP(r.JSON(), remoteAddr)
			//if err != nil {
			//	return err
			//}
		}
	}
}
