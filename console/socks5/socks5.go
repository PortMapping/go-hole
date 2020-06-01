package main

import (
	"fmt"

	"github.com/goextension/log/zap"
	"github.com/portmapping/lurker/nat"
	"github.com/portmapping/lurker/proxy"
)

func main() {
	zap.InitZapSugar()
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
	fmt.Println("listen on port", 10080)
	for {
		accept, err := l.Accept()
		if err != nil {
			continue
		}
		go px.Monitor(accept)
	}

}
