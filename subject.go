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
	connector.ConnectorListener().ID(func(id string) {
		s.connectors.Store(id, connector)
	})

	return nil
}

// NewSubject ...
func NewSubject() Subject {
	return &subject{}
}
