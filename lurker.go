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
	Ping()
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

// JSON ...
func (r ListenResponse) JSON() []byte {
	marshal, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return marshal
}
