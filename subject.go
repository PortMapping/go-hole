package lurker

import (
	"sync"
)

// Subject ...
type Subject interface {
	Add(connector Connector) error
}

type subject struct {
	connectors sync.Map
}

// Add ...
func (s *subject) Add(connector Connector) error {
	s.connectors.Store(connector.ID(), connector)
	return nil
}

// NewSubject ...
func NewSubject() Subject {
	return &subject{}
}
