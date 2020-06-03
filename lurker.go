package lurker

import (
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"

	"github.com/portmapping/lurker/common"
	"github.com/portmapping/lurker/nat"
)

const maxByteSize = 65520

// ListenResponse ...
type ListenResponse struct {
	Status int
	Addr   common.Addr
	Error  error
}

// Lurker ...
type Lurker interface {
	Listen() (c <-chan Connector, err error)
	ListenOnMonitor() error
	RegisterListener(name string, listener Listener)
	Listener(name string) (Listener, bool)
	NetworkNAT(name string) nat.NAT
	Config() Config
}

type lurker struct {
	listeners  map[string]Listener
	cfg        *Config
	timeout    time.Duration
	connectors chan Connector
	pool       *ants.Pool
}

// ListenNoMonitor ...
func (l *lurker) ListenOnMonitor() error {
	connectors, err := l.Listen()
	if err != nil {
		return err
	}
	for connector := range connectors {
		var id string
		var addrs common.Addr
		l := connector.ConnectorListener()
		wg := sync.WaitGroup{}
		wg.Add(2)
		l.ID(func(s string) {
			wg.Done()
			id = s
		})
		l.Addr(func(addr common.Addr) {
			wg.Done()
			addrs = addr
		})
		go func() {
			wg.Wait()
			fmt.Println("id", id, "address is", addrs.String())
		}()
	}
	return nil
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

// Listener ...
func (l *lurker) Listener(name string) (lis Listener, b bool) {
	lis, b = l.listeners[name]
	return
}

// PortUDP ...
func (l *lurker) Config() Config {
	return *l.cfg
}

// Stop ...
func (l *lurker) Stop() error {
	for _, listener := range l.listeners {
		err := listener.Stop()
		if err != nil {
			return err
		}
	}

	fmt.Println("stopped")
	return nil
}

// New ...
func New(cfg *Config) Lurker {
	pool, err := ants.NewPool(5000)
	if err != nil {
		panic(err)
	}
	o := &lurker{
		cfg:        cfg,
		listeners:  make(map[string]Listener),
		connectors: make(chan Connector, 5),
		timeout:    DefaultTimeout,
		pool:       pool,
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
func (l *lurker) Listen() (c <-chan Connector, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Errorw("listener error found", "error", e)
			l.Stop()
		}
	}()

	var lis []string
	for name, listener := range l.listeners {
		lis = append(lis, name)
		go listener.Listen(l.connectors)

		if l.cfg.NAT {
			if v, b := listener.(MappingListener); b {
				err := v.NAT().Mapping()
				if err != nil {
					return nil, err
				}
			}
		}

	}
	l.waitingForReady()

	return l.connectors, nil
}

// JSON ...
func (r ListenResponse) JSON() []byte {
	marshal, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return marshal
}
