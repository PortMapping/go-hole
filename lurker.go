package lurker

import (
	"context"
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
}

// Lurker ...
type Lurker interface {
	Listener
	Stop() error
	NAT() nat.NAT
	PortUDP() int
	PortHole() int
	PortTCP() int
}

type lurker struct {
	ctx         context.Context
	cancel      context.CancelFunc
	udpListener *net.UDPConn
	tcpListener net.Listener
	nat         nat.NAT
	udpPort     int
	holePort    int
	tcpPort     int
	//isMapping   bool
	//mappingPort int
	client  chan Source
	timeout time.Duration
}

// PortUDP ...
func (l *lurker) PortUDP() int {
	return l.udpPort
}

// PortHole ...
func (l *lurker) PortHole() int {
	return l.holePort
}

// PortTCP ...
func (l *lurker) PortTCP() int {
	return l.tcpPort
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
func New() Lurker {
	o := &lurker{
		client:  make(chan Source, 5),
		udpPort: DefaultUDP,
		tcpPort: DefaultTCP,
		timeout: time.Duration(DefaultTimeout),
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
	udpAddr := LocalUDPAddr(l.udpPort)
	fmt.Println("listen udp on address:", udpAddr.String())
	l.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	go listenUDP(l.ctx, l.udpListener, l.client)

	l.nat, err = nat.FromLocal(l.tcpPort)
	tcpAddr := LocalTCPAddr(l.tcpPort)

	if err != nil {
		log.Debugw("nat error", "error", err)
		if err == p2pnat.ErrNoNATFound {
			fmt.Println("listen tcp on address:", tcpAddr.String())
			l.tcpListener, err = reuse.ListenTCP("tcp", tcpAddr)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		//l.isMapping = true
		extPort, err := l.nat.Mapping()
		if err != nil {
			return nil, err
		}
		address, err := l.nat.GetInternalAddress()
		if err != nil {
			return nil, err
		}
		addr := ParseSourceAddr("tcp", address, extPort)
		fmt.Println("mapping on address:", addr.String())
		l.tcpListener, err = reuse.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return nil, err
		}
		l.holePort = extPort
	}
	go listenTCP(l.ctx, l.tcpListener, l.client)

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

			service, err := ParseService(data[:n])
			if err != nil {
				//waiting for next
				log.Debugw("debug|listenUDP|ParseService", "error", err)
				continue
			}
			c := source{
				addr: Addr{
					Protocol: remoteAddr.Network(),
					Port:     remoteAddr.Port,
					IP:       remoteAddr.IP,
				},
				service: service,
			}
			cli <- &c
			_, err = listener.WriteToUDP([]byte(c.addr.String()), remoteAddr)
			if err != nil {
				return err
			}
		}
	}
}

//// IsMapping ...
//func (l *lurker) IsMapping() bool {
//	return l.isMapping
//}
//
//// MappingPort ...
//func (l *lurker) MappingPort() int {
//	return l.mappingPort
//}

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
		service, err := ParseService(data[:n])
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
		_, err = conn.Write([]byte(c.addr.String()))
		if err != nil {
			log.Debugw("debug|getClientFromTCP|write", "error", err)
			return err
		}
	}
	return nil
}
