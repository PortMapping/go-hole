package lurker

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/portmapping/go-reuse"
	"github.com/portmapping/lurker/nat"

	p2pnat "github.com/libp2p/go-nat"
)

const maxByteSize = 65520

// Lurker ...
type Lurker interface {
	Listener() (c <-chan Source, err error)
	Stop() error
	IsMapping() bool
	MappingPort() int
}

type lurker struct {
	ctx         context.Context
	cancel      context.CancelFunc
	udpListener *net.UDPConn
	tcpListener net.Listener
	nat         nat.NAT
	udpPort     int
	tcpPort     int
	isMapping   bool
	mappingPort int
	client      chan Source
	timeout     time.Duration
}

// Stop ...
func (o *lurker) Stop() error {
	if err := o.nat.StopMapping(); err != nil {
		return err
	}
	if o.cancel != nil {
		o.cancel()
		o.cancel = nil
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

// Listener ...
func (o *lurker) Listener() (c <-chan Source, err error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("listener error found", e)
		}
	}()
	o.udpListener, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: o.udpPort})
	if err != nil {
		return nil, err
	}

	go listenUDP(o.ctx, o.udpListener, o.client)

	o.nat, err = nat.FromLocal(o.tcpPort)
	if err != nil {
		if err == p2pnat.ErrNoNATFound {
			o.tcpListener, err = reuse.Listen("tcp", LocalAddr(o.tcpPort))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		o.isMapping = true
		extPort, err := o.nat.Mapping()
		if err != nil {
			return nil, err
		}
		o.tcpListener, err = reuse.Listen("tcp", LocalAddr(extPort))
		if err != nil {
			return nil, err
		}
		o.mappingPort = extPort
	}
	go listenTCP(o.ctx, o.tcpListener, o.client)

	return o.client, nil
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
				continue
			}
			log.Printf("<%s> %s\n", remoteAddr.String(), data[:n])
			c := source{
				addr: Addr{
					Network: remoteAddr.Network(),
					Address: remoteAddr.String(),
				},
				data: make([]byte, n),
			}
			copy(c.data, data[:n])
			cli <- &c
			_, err = listener.Write(c.addr.JSON())
			if err != nil {
				return err
			}
		}
	}
}

// IsMapping ...
func (o *lurker) IsMapping() bool {
	return o.isMapping
}

// MappingPort ...
func (o *lurker) MappingPort() int {
	return o.mappingPort
}

func listenTCP(ctx context.Context, listener net.Listener, cli chan<- Source) (err error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			acceptTCP, err := listener.Accept()
			if err != nil {
				continue
			}
			go getClientFromTCP(ctx, acceptTCP, cli)
		}
	}
}

func getClientFromTCP(ctx context.Context, conn net.Conn, cli chan<- Source) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		data := make([]byte, maxByteSize)
		n, err := conn.Read(data)
		if err != nil {
			return err
		}
		log.Printf("<%s> %s\n", conn.RemoteAddr().String(), string(data[:n]))
		c := source{
			addr: Addr{
				Network: conn.RemoteAddr().Network(),
				Address: conn.RemoteAddr().String(),
			},
			data: make([]byte, n),
		}
		copy(c.data, data[:n])
		cli <- &c
		_, err = conn.Write(c.addr.JSON())
		if err != nil {
			return err
		}
	}
	return nil
}
