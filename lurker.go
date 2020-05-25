package lurker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/portmapping/lurker/nat"
)

const maxByteSize = 65520

// NATer ...
type NATer interface {
	IsSupport() bool
	NAT() nat.NAT
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
	RegisterListener(name string, listener Listener)
	Listener(name string) (Listener, bool)
	NetworkNAT(name string) nat.NAT
	NetworkMappingPort(name string) int
	Config() Config
}

type lurker struct {
	listeners map[string]Listener
	cfg       *Config
	nat       nat.NAT
	sources   chan Source
	timeout   time.Duration
}

// NetworkNAT ...
func (l *lurker) NetworkNAT(name string) nat.NAT {
	listener, b := l.listeners[name]
	if b {
		ter, b := listener.(NATer)
		if b && ter.IsSupport() {
			return ter.NAT()
		}
	}
	return nil
}

// NetworkMappingPort ...
func (l *lurker) NetworkMappingPort(name string) int {
	listener, b := l.listeners[name]
	if b {
		return listener.MappingPort()
	}
	return 0
}

// Listener ...
func (l *lurker) Listener(name string) (lis Listener, b bool) {
	lis, b = l.listeners[name]
	return
}

// PortUDP ...
func (l *lurker) Config() Config {
	return *l.cfg
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
		cfg:       cfg,
		sources:   make(chan Source, 5),
		listeners: make(map[string]Listener),
		timeout:   DefaultTimeout,
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

func (l *lurker) waitingForReady() {
	total := len(l.listeners)
	for {
		count := 0
		for _, listener := range l.listeners {
			if listener.IsReady() {
				count++
			}
		}
		if count == total {
			return
		}
		time.Sleep(3 * time.Second)
	}
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
	l.waitingForReady()
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
