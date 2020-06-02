package main

import (
	"fmt"

	"github.com/portmapping/lurker"
	"github.com/spf13/cobra"
)

func cmdServer() *cobra.Command {
	tcp := 0
	//udp := 0
	nat := false
	cmd := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := lurker.DefaultConfig()
			cfg.TCP = tcp
			//cfg.UDP = udp
			cfg.NAT = nat
			l := lurker.New(cfg)
			t := lurker.NewTCPListener(cfg)
			l.RegisterListener("tcp", t)
			err := l.ListenOnMonitor()
			if err != nil {
				panic(err)
				return
			}
			fmt.Println("your connect id:", lurker.GlobalID)
			go func() {
				//for connector := range connectors {
				//	fmt.Println(connector.ID())
				//}
			}()
			waitForSignal()
		},
	}
	cmd.Flags().IntVarP(&tcp, "tcp", "t", 16004, "handle tcp port")
	//cmd.Flags().IntVarP(&udp, "udp", "u", 16005, "handle udp port")
	cmd.Flags().BoolVarP(&nat, "nat", "n", false, "enable nat")
	return cmd
}
