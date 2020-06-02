package lurker

// Connector ...
type Connector interface {
	RegisterCallback(cb ConnectorCallback)
	ID(f func(string))
	Header() (HandshakeHead, error)
	Close() error
	Response(header HandshakeHead) error
}

// ConnectorCallback ...
type ConnectorCallback func(rt RequestType, data []byte)
