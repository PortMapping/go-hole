package lurker

// Listener ...
type Listener interface {
	Listen() (err error)
	MappingPort() int
	Stop() error
	IsReady() bool
	Accept() <-chan Connector
}
