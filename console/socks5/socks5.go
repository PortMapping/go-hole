package main

import (
	"github.com/portmapping/lurker/nat"
	"github.com/portmapping/lurker/proxy"
)

func main() {
	p, err := nat.FromLocal("tcp", 10080)
	if err != nil {
		panic(err)
	}
	px, err := proxy.New("socks5", p, proxy.NoAuth())
	if err != nil {
		panic(err)
	}
	l, err := px.ListenPort(10080)
	if err != nil {
		panic(err)
	}
	for {
		accept, err := l.Accept()
		if err != nil {
			continue
		}
		go px.Monitor(accept)
	}

}
