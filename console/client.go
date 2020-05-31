package main

import (
	"fmt"
	"net"

	"github.com/portmapping/lurker"
	"github.com/portmapping/lurker/common"
	"github.com/spf13/cobra"
)

func cmdClient() *cobra.Command {
	var addr string
	var local int
	var network string
	var proxy string
	var proxyPort int
	var proxyName string
	var proxyPass string
	cmd := &cobra.Command{
		Use: "client",
		Run: func(cmd *cobra.Command, args []string) {
			addrs, i := common.ParseAddr(addr)
			localAddr := net.IPv4zero
			ispAddr := net.IPv4zero

			cfg := lurker.DefaultConfig()
			cfg.Proxy = append(cfg.Proxy, lurker.Proxy{
				Type: proxy,
				Nat:  true,
				Port: proxyPort,
				Name: proxyName,
				Pass: proxyPass,
			})
			l := lurker.New(cfg)

			_, err := lurker.RegisterLocalProxy(l, cfg)
			if err != nil {
				panic(err)
			}

			go l.ListenMonitor()
			fmt.Println("remote addr:", addrs.String(), i)
			s := lurker.NewSource(lurker.Service{
				ID:    lurker.GlobalID,
				ISP:   ispAddr,
				Local: localAddr,
			}, common.Addr{
				Protocol: network,
				IP:       addrs,
				Port:     i,
			})
			s.SetMappingPort("tcp", 10080)
			err = s.Connect()
			if err != nil {
				panic(err)
			}
			fmt.Println("target connected")
			waitForSignal()
		},
	}
	cmd.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:16004", "default 127.0.0.1:16004")
	cmd.Flags().StringVarP(&network, "network", "n", "tcp", "")
	cmd.Flags().IntVarP(&local, "local", "l", 16004, "handle local mapping port")
	cmd.Flags().StringVarP(&proxy, "proxy", "p", "socks5", "locak proxy")
	cmd.Flags().StringVarP(&proxyName, "pname", "", "", "local proxy port")
	cmd.Flags().StringVarP(&proxyPass, "ppass", "", "", "local proxy port")
	cmd.Flags().IntVarP(&proxyPort, "pport", "", 10080, "local proxy port")
	return cmd
}
