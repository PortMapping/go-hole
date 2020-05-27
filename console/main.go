package main

import (
	"fmt"
	address2 "github.com/portmapping/lurker/common"
	"github.com/portmapping/lurker/nat"
	"net"
	"os"
	"strconv"
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
	listen := true
	if len(os.Args) > 3 {
		parseBool, err := strconv.ParseBool(os.Args[3])
		if err != nil {
			listen = true
		} else {
			listen = parseBool
		}

	}

	cfg := lurker.DefaultConfig()
	cfg.TCP = 16004
	//cfg.UDP = 16005

	l := lurker.New(cfg)
	//if l.NAT() != nil {
	//	localAddr, _ = l.NAT().GetInternalAddress()
	//	ispAddr, _ = l.NAT().GetExternalAddress()
	//}
	err := lurker.RegisterLocalProxy(l, cfg)
	if err != nil {
		return
	}
	tcpPort, udpPort := 0, 0
	if listen {
		fmt.Println("listen with port mapping")
		t := lurker.NewTCPListener(cfg)
		//u := lurker.NewUDPListener(cfg)
		l.RegisterListener("tcp", t)
		//l.RegisterListener("udp", u)
		connectors, err := l.Listen()
		if err != nil {
			panic(err)
			return
		}

		fmt.Println("your connect id:", lurker.GlobalID)
		go func() {
			for connector := range connectors {
				//fmt.Println("connect from:", source.Addr().String(), string(source.Service().JSON()))
				//_, ok := list.Load(source.Service().ID)
				//if ok {
				//	fmt.Println("exist:", source.Service().ID)
				//	continue
				//}
				//s := lurker.NewSource(lurker.Service{
				//	ID:          lurker.GlobalID,
				//	ISP:         ispAddr,
				//	Local:       localAddr,
				//	PortUDP:     l.Config().UDP,
				//	PortTCP:     l.Config().TCP,
				//	KeepConnect: false,
				//}, source.Addr())
				//go func(id string, s lurker.Source) {
				//	err := s.Try()
				//	fmt.Println("reverse connected:", err)
				//	if err != nil {
				//		return
				//	}
				//	list.Store(id, s)
				//}(source.Service().ID, s)
				fmt.Println(connector.ID())
			}
		}()
	} else {
		nat, err := nat.FromLocal("tcp", l.Config().TCP)
		if err != nil {
			fmt.Println("error tcpport", err)
		}
		err = nat.Mapping()
		if err != nil {
			fmt.Println("error tcpport", err)
		}
		tcpPort = nat.ExtPort()
	}
	if len(os.Args) > 2 {
		addr, i := address2.ParseAddr(address)
		localAddr := net.IPv4zero
		ispAddr := net.IPv4zero

		fmt.Println("remote common:", addr.String(), i)
		s := lurker.NewSource(lurker.Service{
			ID:      lurker.GlobalID,
			ISP:     ispAddr,
			Local:   localAddr,
			PortUDP: l.Config().UDP,
			PortTCP: l.Config().TCP,
		}, address2.Addr{
			Protocol: network,
			IP:       addr,
			Port:     i,
		})
		s.SetMappingPort("tcp", tcpPort)
		s.SetMappingPort("udp", udpPort)
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
