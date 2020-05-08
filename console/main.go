package main

import (
	"fmt"
	"github.com/portmapping/lurker"
	"os"
	"time"
)

func main() {
	lurker.DefaultTCP = 16004
	lurker.DefaultUDP = 16005
	network := "tcp"
	address := ""
	msg := "hello world"
	if len(os.Args) > 2 {
		network = os.Args[1]
		address = os.Args[2]
	}
	if len(os.Args) > 3 {
		msg = os.Args[3]
	}
	l := lurker.New()
	listener, err := l.Listener()
	if err != nil {
		return
	}
	go func() {
		for source := range listener {
			b := source.Ping(msg)
			fmt.Println("source connected:", b)
		}
	}()
	if len(os.Args) > 2 {
		s := lurker.NewSource(network, address)
		fmt.Println("target connected:", s.Ping(msg))
	}
	time.Sleep(30 * time.Minute)
}
