package lurker

import (
	"context"
	"fmt"
	"net"
)

type udpListener struct {
	ctx         context.Context
	cancel      context.CancelFunc
	source      chan Source
	port        int
	mappingPort int
	nat         bool //unused now
	udpListener *net.UDPConn
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
		nat:    cfg.NAT,
		port:   cfg.UDP,
		source: make(chan Source),
	}
	udp.ctx, udp.cancel = context.WithCancel(context.TODO())
	return udp
}

// Listen ...
func (l *udpListener) Listen() (c <-chan Source, err error) {
	udpAddr := LocalUDPAddr(l.port)
	fmt.Println("listen udp on address:", udpAddr.String())
	l.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	go listenUDP(l.ctx, l.udpListener, l.source)

	return l.source, nil
}

func listenUDP(ctx context.Context, listener *net.UDPConn, cli chan<- Source) (err error) {
	data := make([]byte, maxByteSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, remoteAddr, err := listener.ReadFromUDP(data)
			if err != nil {
				//waiting for next
				log.Debugw("debug|listenUDP|ReadFromUDP", "error", err)
				continue
			}
			//handshake, err := ParseHandshake(data)
			//if err != nil {
			//	return err
			//}

			service, err := DecodeHandshakeRequest(data[:n])
			if err != nil {
				//waiting for next
				log.Debugw("debug|listenUDP|ParseService", "error", err)
				continue
			}

			netAddr := ParseNetAddr(remoteAddr)
			c := source{
				addr:    *netAddr,
				service: service,
			}
			cli <- &c
			err = tryReverseUDP(&source{
				addr: *netAddr,
				service: Service{
					ID:          GlobalID,
					KeepConnect: false,
				}})
			status := 0
			if err != nil {
				status = -1
				log.Debugw("debug|listenUDP|tryReverseUDP", "error", err)
			}

			r := &ListenResponse{
				Status: status,
				Addr:   *netAddr,
				Error:  err,
			}
			_, err = listener.WriteToUDP(r.JSON(), remoteAddr)
			if err != nil {
				return err
			}
		}
	}
}
