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
	cfg.TCP = 16004
	cfg.UDP = 16005

	l := lurker.New(cfg)
	localAddr := net.IPv4zero
	ispAddr := net.IPv4zero
	//if l.NAT() != nil {
	//	localAddr, _ = l.NAT().GetInternalAddress()
	//	ispAddr, _ = l.NAT().GetExternalAddress()
	//}
	t := lurker.NewTCPListener(cfg)
	u := lurker.NewUDPListener(cfg)
	l.RegisterListener("tcp", t)
	l.RegisterListener("udp", u)
	listener, err := l.Listen()
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("your connect id:", lurker.GlobalID)
	go func() {
		for source := range listener {
			fmt.Println("connect from:", source.Addr().String(), string(source.Service().JSON()))
			_, ok := list.Load(source.Service().ID)
			if ok {
				fmt.Println("exist:", source.Service().ID)
				continue
			}
			s := lurker.NewSource(lurker.Service{
				ID:          lurker.GlobalID,
				ISP:         ispAddr,
				Local:       localAddr,
				PortUDP:     l.Config().UDP,
				PortTCP:     l.Config().TCP,
				KeepConnect: false,
			}, source.Addr())
			s.SetMappingPort(network, l.NetworkMappingPort(network))
			go func(id string, s lurker.Source) {
				err := s.Try()
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
		//if l.NAT() != nil {
		//	localAddr, _ = l.NAT().GetInternalAddress()
		//	ispAddr, _ = l.NAT().GetExternalAddress()
		//}
		fmt.Println("remote addr:", addr.String(), i)
		s := lurker.NewSource(lurker.Service{
			ID:      lurker.GlobalID,
			ISP:     ispAddr,
			Local:   localAddr,
			PortUDP: l.Config().UDP,
			PortTCP: l.Config().TCP,
		}, lurker.Addr{
			Protocol: network,
			IP:       addr,
			Port:     i,
		})
		s.SetMappingPort(network, l.NetworkMappingPort(network))
		go func() {
			_, ok := list.Load(s.Service().ID)
			if ok {
				return
			}
			err := s.Connect()
			fmt.Println("target connected:", err)
			if err != nil {
				list.Store(s.Service().ID, s)
			}
		}()

	}
	fmt.Println("ready for waiting")
	time.Sleep(30 * time.Minute)
}
