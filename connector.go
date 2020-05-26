package lurker

import "io"

// Connector ...
type Connector interface {
	ID() string
	Process()
	RegisterCallback(cb ConnectorCallback)
}

// ConnectorCallback ...
type ConnectorCallback func(rt RequestType, closer io.ReadWriteCloser)
