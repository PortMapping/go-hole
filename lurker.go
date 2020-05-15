package lurker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/nat"

	p2pnat "github.com/libp2p/go-nat"
)

const maxByteSize = 65520

// Listener ...
type Listener interface {
	Listen() (c <-chan Source, err error)
	Stop() error
}

// ListenResponse ...
type ListenResponse struct {
	Status int
	Addr   Addr
	Error  error
}

// Lurker ...
type Lurker interface {
	Listener
	NAT() nat.NAT
	Config() Config
	PortHole() int
}

type lurker struct {
	ctx         context.Context
	cancel      context.CancelFunc
	udpListener *net.UDPConn
	tcpListener net.Listener
	cfg         *Config
	nat         nat.NAT
	holePort    int
	client      chan Source
	timeout     time.Duration
}

// PortUDP ...
func (l *lurker) Config() Config {
	return *l.cfg
}

// PortHole ...
func (l *lurker) PortHole() int {
	return l.holePort
}

// NAT ...
func (l *lurker) NAT() nat.NAT {
	return l.nat
}

// Stop ...
func (l *lurker) Stop() error {
	if err := l.nat.StopMapping(); err != nil {
		return err
	}
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	return nil
}

// New ...
func New(cfg *Config) Lurker {
	o := &lurker{
		cfg:     cfg,
		client:  make(chan Source, 5),
		timeout: DefaultTimeout,
	}
	o.ctx, o.cancel = context.WithCancel(context.TODO())
	return o
}

// Listen ...
func (l *lurker) Listen() (c <-chan Source, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Errorw("listener error found", "error", e)
		}
	}()

	if l.cfg.UDP != 0 {
		udpAddr := LocalUDPAddr(l.cfg.UDP)
		fmt.Println("listen udp on address:", udpAddr.String())
		l.udpListener, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, err
		}

		go listenUDP(l.ctx, l.udpListener, l.client)
	}

	if l.cfg.TCP != 0 {
		tcpAddr := LocalTCPAddr(l.cfg.TCP)
		if l.cfg.Secret != nil {
			l.tcpListener, err = reuse.ListenTLS("tcp", DefaultLocalTCPAddr.String(), l.cfg.Secret)
		} else {
			l.tcpListener, err = reuse.ListenTCP("tcp", tcpAddr)
		}
		if err != nil {
			return nil, err
		}
		go listenTCP(l.ctx, l.tcpListener, l.client)

		if !l.cfg.NAT {
			return l.client, nil
		}

		l.nat, err = nat.FromLocal(l.cfg.TCP)
		if err != nil {
			log.Debugw("nat error", "error", err)
			if err == p2pnat.ErrNoNATFound {
				fmt.Println("listen tcp on address:", tcpAddr.String())
			}
			l.cfg.NAT = false
		} else {
			extPort, err := l.nat.Mapping()
			if err != nil {
				log.Debugw("nat mapping error", "error", err)
				l.cfg.NAT = false
				return l.client, nil
			}
			l.holePort = extPort

			address, err := l.nat.GetExternalAddress()
			if err != nil {
				log.Debugw("get external address error", "error", err)
				l.cfg.NAT = false
				return l.client, nil
			}
			addr := ParseSourceAddr("tcp", address, extPort)
			fmt.Println("mapping on address:", addr.String())
		}
	}
	return l.client, nil
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
		data := make([]byte, maxByteSize)
		n, err := conn.Read(data)
		if err != nil {
			log.Debugw("debug|getClientFromTCP|Read", "error", err)
			return err
		}
		ip, port := ParseAddr(conn.RemoteAddr().String())
		service, err := DecodeHandshakeRequest(data[:n])
		if err != nil {
			log.Debugw("debug|getClientFromTCP|ParseService", "error", err)
			return err
		}
		if service.KeepConnect {
			close = false
		}
		c := source{
			addr: Addr{
				Protocol: conn.RemoteAddr().Network(),
				IP:       ip,
				Port:     port,
			},
			service: service,
		}
		cli <- &c
		netAddr := ParseNetAddr(conn.RemoteAddr())

		err = tryReverseTCP(&source{addr: *netAddr,
			service: Service{
				ID:          GlobalID,
				KeepConnect: false,
			}})
		status := 0
		if err != nil {
			status = -1
			log.Debugw("debug|getClientFromTCP|tryReverseTCP", "error", err)
		}

		r := &ListenResponse{
			Status: status,
			Addr:   *netAddr,
			Error:  err,
		}
		_, err = conn.Write(r.JSON())
		if err != nil {
			log.Debugw("debug|getClientFromTCP|write", "error", err)
			return err
		}
	}
	return nil
}

// JSON ...
func (r ListenResponse) JSON() []byte {
	marshal, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return marshal
}
