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

	address := ""
	if len(os.Args) > 2 {
		address = os.Args[2]
	}
	if len(os.Args) > 3 {
	}
	l := lurker.New()
	listener, err := l.Listen()
	if err != nil {
		panic(err)
		return
	}
	go func() {
		for source := range listener {
			go func(s lurker.Source) {
				b := s.TryConnect()
				fmt.Println("reverse connected:", b)
			}(source)
		}
	}()

	if len(os.Args) > 2 {
		addr, i := lurker.ParseAddr(address)

		internalAddress, err := l.NAT().GetInternalAddress()
		if err != nil {
			return
		}
		fmt.Println("remote addr:", addr.String(), i)
		s := lurker.NewSource(lurker.Service{
			ID:       "random_str",
			ISP:      internalAddress,
			PortUDP:  l.PortUDP(),
			PortHole: l.PortHole(),
			PortTCP:  l.PortTCP(),
			ExtData:  nil,
		})
		go func() {
			b := s.TryConnect()
			fmt.Println("target connected:", b)
		}()

	}
	fmt.Println("ready for waiting")
	time.Sleep(30 * time.Minute)
}
