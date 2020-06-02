package lurker

// Connector ...
type Connector interface {
	RegisterCallback(cb ConnectorCallback)
	ID(f func(string))
	Header() (HandshakeHead, error)
	Close() error
	Reply(status HandshakeStatus, data []byte) error
	Do(handshakeType HandshakeType) error
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
		log.Debugw("connector get header", "error", err)
		e := connector.Reply(HandshakeStatusFailed, nil)
		if e != nil {
			log.Debugw("connector response", "error", e)
		}
		return err
	}

	return connector.Do(header.Type)

}
