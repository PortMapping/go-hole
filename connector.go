package lurker

// Connector ...
type Connector interface {
	Process()
	RegisterCallback(cb ConnectorCallback)
	ID(f func(string))
}

// ConnectorCallback ...
type ConnectorCallback func(rt RequestType, data []byte)
