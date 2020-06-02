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

func receive(connector Connector) (err error) {
	defer func() {
		if err != nil {
			connector.Close()
		}
	}()
	header, err := connector.Header()
	if err != nil {
		return err
	}
	err = connector.Response(header)
	return
}
