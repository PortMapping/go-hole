package main

import "github.com/spf13/cobra"

func cmdServer() *cobra.Command {
	tcp := 0
	udp := 0
	cmd := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
	cmd.Flags().IntVarP(&tcp, "tcp", "t", 16004, "handle tcp port")
	cmd.Flags().IntVarP(&udp, "udp", "u", 16005, "handle udp port")
	return cmd
}
