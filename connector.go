package lurker

// Connector ...
type Connector interface {
	RegisterCallback(cb ConnectorCallback)
	ID(f func(string))
	Header() (HandshakeHead, error)
	Close() error
	Reply(header HandshakeHead, status HandshakeStatus, data []byte) error
	Pong() error
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
		e := connector.Reply(header, HandshakeStatusFailed, nil)
		if e != nil {
			log.Debugw("connector response", "error", e)
		}
		return err
	}
	switch header.Type {
	case HandshakeTypePing:
		return connector.Pong()
	case HandshakeTypeConnect:
		return connector.Interaction()
	case HandshakeTypeAdapter:
		return connector.Intermediary()
	}
	return connector.Other()
}
