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
	cmd := &cobra.Command{
		Use: "client",
		Run: func(cmd *cobra.Command, args []string) {
			addrs, i := common.ParseAddr(addr)
			localAddr := net.IPv4zero
			ispAddr := net.IPv4zero

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

			err := s.Connect()
			fmt.Println("target connected:", err)
		},
	}
	cmd.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:16004", "default 127.0.0.1:16004")
	cmd.Flags().StringVarP(&network, "network", "n", "tcp", "")
	cmd.Flags().IntVarP(&local, "local", "l", 16004, "handle local mapping port")
	cmd.Flags().StringVarP(&proxy, "proxy", "p", "socks5", "locak proxy")

	cmd.Flags().IntVarP(&proxyPort, "proxy_port", "pp", 10080, "local proxy port")
	return cmd
}
