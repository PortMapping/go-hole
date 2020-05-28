package main

import (
	"fmt"

	"github.com/portmapping/lurker"
	"github.com/spf13/cobra"
)

func cmdServer() *cobra.Command {
	tcp := 0
	udp := 0
	cmd := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := lurker.DefaultConfig()
			cfg.TCP = tcp
			cfg.UDP = udp

			l := lurker.New(cfg)
			t := lurker.NewTCPListener(cfg)
			l.RegisterListener("tcp", t)
			connectors, err := l.Listen()
			if err != nil {
				panic(err)
				return
			}

			fmt.Println("your connect id:", lurker.GlobalID)
			go func() {
				for connector := range connectors {
					fmt.Println(connector.ID())
				}
			}()
		},
	}
	cmd.Flags().IntVarP(&tcp, "tcp", "t", 16004, "handle tcp port")
	cmd.Flags().IntVarP(&udp, "udp", "u", 16005, "handle udp port")
	return cmd
}
