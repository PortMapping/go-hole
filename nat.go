package natpmp

type NAT interface {
	Mapping() (port int, err error)
	StopMapping()
	//GetExternalAddress() (addr net.IP, err error)
	//AddPortMapping(protocol string, externalPort, internalPort int) (mappedExternalPort int, err error)
	//DeletePortMapping(protocol string, externalPort, internalPort int) (err error)
}
