package main

import "github.com/spf13/cobra"

func cmdClient() *cobra.Command {
	var addr string
	var local int
	var network string
	cmd := &cobra.Command{
		Use: "client",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
	cmd.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:16004", "default 127.0.0.1:16004")
	cmd.Flags().StringVarP(&network, "network", "n", "tcp", "")
	cmd.Flags().IntVarP(&local, "local", "l", 16004, "handle local mapping port")
	return cmd
}
