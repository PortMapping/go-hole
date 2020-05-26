package lurker

// Connector ...
type Connector interface {
	ID() string
	Process()
}
