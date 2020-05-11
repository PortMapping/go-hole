package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/portmapping/lurker"
)

func main() {
	lurker.DefaultTCP = 16004
	lurker.DefaultUDP = 16005

	address := ""
	list := sync.Map{}
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
	fmt.Println("your connect id:", lurker.GlobalID)
	go func() {
		for source := range listener {
			tmp := source
			go func(s lurker.Source) {
				fmt.Println("connect from:", s.Addr().String(), string(s.Service().JSON()), string(s.Service().ExtData))
				_, ok := list.Load(s.Service().ID)
				if ok {
					return
				}
				localAddr := net.IPv4zero
				ispAddr := net.IPv4zero
				if l.NAT() != nil {
					localAddr, _ = l.NAT().GetInternalAddress()
					ispAddr, _ = l.NAT().GetExternalAddress()
				}
				source := lurker.NewSource(lurker.Service{
					ID:       lurker.GlobalID,
					ISP:      ispAddr,
					Local:    localAddr,
					PortUDP:  l.PortUDP(),
					PortHole: l.PortHole(),
					PortTCP:  l.PortTCP(),
					ExtData:  nil,
				}, s.Addr())
				err := source.TryConnect()
				fmt.Println("reverse connected:", err)
				if err != nil {
					list.Store(s.Service().ID, source)
				}
			}(tmp)
		}
	}()

	if len(os.Args) > 2 {
		addr, i := lurker.ParseAddr(address)
		extAddr, err := l.NAT().GetExternalAddress()
		localAddr, err := l.NAT().GetInternalAddress()
		if err != nil {
			return
		}
		fmt.Println("remote addr:", addr.String(), i)

		s := lurker.NewSource(lurker.Service{
			ID:       lurker.GlobalID,
			ISP:      extAddr,
			Local:    localAddr,
			PortUDP:  l.PortUDP(),
			PortHole: l.PortHole(),
			PortTCP:  l.PortTCP(),
			ExtData:  nil,
		}, lurker.Addr{
			Protocol: "tcp",
			IP:       addr,
			Port:     i,
		})
		go func() {
			_, ok := list.Load(s.Service().ID)
			if ok {
				return
			}
			err := s.TryConnect()
			fmt.Println("target connected:", err)
			if err != nil {
				list.Store(s.Service().ID, s)
			}
		}()

	}
	fmt.Println("ready for waiting")
	time.Sleep(30 * time.Minute)
}
