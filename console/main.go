package main

import (
	"fmt"
	"github.com/portmapping/lurker"
	"os"
	"time"
)

func main() {
	network := "tcp"
	address := ""
	if len(os.Args) > 2 {
		network = os.Args[1]
		address = os.Args[2]
	}
	l := lurker.New()
	listener, err := l.Listener()
	if err != nil {
		return
	}
	go func() {
		for source := range listener {
			b := source.Ping()
			fmt.Println("source connected:", b)
		}
	}()

	s := lurker.NewSource(network, address)
	fmt.Println("target connected:", s.Ping())

	time.Sleep(30 * time.Second)

}
