package lurker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/portmapping/lurker/nat"
)

const maxByteSize = 65520

// Listener ...
type Listener interface {
	Listen(chan<- Source) (err error)
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
	Listen() (c <-chan Source, err error)
	NAT() nat.NAT
	Config() Config
	PortHole() int
}

type lurker struct {
	listeners map[string]Listener

	tcpListener net.Listener
	cfg         *Config
	nat         nat.NAT
	holePort    int
	sources     chan Source
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
	fmt.Println("stopped")
	return nil
}

// New ...
func New(cfg *Config) Lurker {
	o := &lurker{
		cfg:     cfg,
		sources: make(chan Source, 5),
		timeout: DefaultTimeout,
	}
	return o
}

// RegisterListener ...
func (l *lurker) RegisterListener(name string, listener Listener) {
	if name == "" {
		name = UUID()
	}
	l.listeners[name] = listener
}

// Listen ...
func (l *lurker) Listen() (c <-chan Source, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Errorw("listener error found", "error", e)
		}
	}()

	for _, listener := range l.listeners {
		go listener.Listen(l.sources)
	}
	return l.sources, nil
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
