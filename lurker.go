package lurker

import (
	"context"
	"fmt"
	"github.com/libp2p/go-nat"
	"log"
	"net"
	"time"
)

const maxByteSize = 65520

// Lurker ...
type Lurker interface {
	Listener() (c <-chan Source, err error)
	Stop() error
}

type lurker struct {
	ctx         context.Context
	cancel      context.CancelFunc
	udpListener *net.UDPConn
	tcpListener *net.TCPListener
	udpPort     int
	tcpPort     int
	client      chan Source
	timeout     time.Duration
}

// Stop ...
func (o *lurker) Stop() error {
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

	gateway, err := nat.DiscoverGateway()
	if err != nil {
		if err == nat.ErrNoNATFound {
			o.tcpListener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: o.tcpPort})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		extPort, err := gateway.AddPortMapping("tcp", o.tcpPort, "http", o.timeout)
		if err != nil {
			return nil, err
		}
		go keepMapping(o.ctx, gateway, o.tcpPort, o.timeout)
		o.tcpListener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: extPort})
		if err != nil {
			return nil, err
		}
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
			cli <- c
			_, err = listener.Write(c.addr.JSON())
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
		cli <- c
		_, err = conn.Write(c.addr.JSON())
		if err != nil {
			return err
		}
	}
	return nil
}

// KeepMapping ...
func keepMapping(ctx context.Context, n nat.NAT, port int, timeout time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(30 * time.Second)
			_, err := n.AddPortMapping("tcp", port, "mapping", timeout)
			if err != nil {
				return err
			}
		}

	}
}
