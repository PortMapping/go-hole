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
	network := "tcp"
	address := ""
	list := sync.Map{}
	if len(os.Args) > 2 {
		network = os.Args[1]
		address = os.Args[2]
	}
	if len(os.Args) > 3 {
	}

	cfg := lurker.DefaultConfig()

	l := lurker.New(cfg)
	localAddr := net.IPv4zero
	ispAddr := net.IPv4zero
	if l.NAT() != nil {
		localAddr, _ = l.NAT().GetInternalAddress()
		ispAddr, _ = l.NAT().GetExternalAddress()
	}
	listener, err := l.Listen()
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("your connect id:", lurker.GlobalID)
	go func() {
		for source := range listener {
			fmt.Println("connect from:", source.Addr().String(), string(source.Service().JSON()), string(source.Service().ExtData))
			_, ok := list.Load(source.Service().ID)
			if ok {
				fmt.Println("exist:", source.Service().ID)
				continue
			}
			s := lurker.NewSource(lurker.Service{
				ID:          lurker.GlobalID,
				ISP:         ispAddr,
				Local:       localAddr,
				PortUDP:     l.PortUDP(),
				PortHole:    l.PortHole(),
				PortTCP:     l.PortTCP(),
				KeepConnect: false,
				ExtData:     nil,
			}, source.Addr())
			go func(id string, s lurker.Source) {
				err := s.TryConnect()
				fmt.Println("reverse connected:", err)
				if err != nil {
					return
				}
				list.Store(id, s)
			}(source.Service().ID, s)
		}
	}()

	if len(os.Args) > 2 {
		addr, i := lurker.ParseAddr(address)
		localAddr := net.IPv4zero
		ispAddr := net.IPv4zero
		if l.NAT() != nil {
			localAddr, _ = l.NAT().GetInternalAddress()
			ispAddr, _ = l.NAT().GetExternalAddress()
		}
		fmt.Println("remote addr:", addr.String(), i)

		s := lurker.NewSource(lurker.Service{
			ID:       lurker.GlobalID,
			ISP:      ispAddr,
			Local:    localAddr,
			PortUDP:  l.PortUDP(),
			PortHole: l.PortHole(),
			PortTCP:  l.PortTCP(),
			ExtData:  nil,
		}, lurker.Addr{
			Protocol: network,
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
