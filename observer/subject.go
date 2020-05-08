package observer

import (
	"crypto/sha256"
	"fmt"
)

// Subject ...
type Subject interface {
}

type subject struct {
	sources map[string]map[string]Source
}

// NewSubject ...
func NewSubject() Subject {
	return subject{
		sources: make(map[string]map[string]Source),
	}
}

// Ping ...
func (s *subject) Ping() {

}

// RegisterSource ...
func (s *subject) RegisterSource(name string, source Source) {
	hash := hashString(source.Network(), source.String())
	if _, b := s.sources[name]; b {
		if _, b := s.sources[name][hash]; b {
			return
		}
		s.sources[name][hash] = source
		return
	}
	s.sources[name] = map[string]Source{
		hash: source,
	}
	return
}

func hashString(network, address string) string {
	s := sha256.Sum256([]byte(network + "_" + address))
	return fmt.Sprintf("%x", s)
}
