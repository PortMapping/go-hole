package main

import (
	"fmt"
	"os/signal"

	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "Lurker",
	Short: "Lurker is a intranet direct connection tool",
	Long:  `Intranet direct connection allows you to directly access the intranet to achieve the fastest access speed.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func waitForSignal() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
}

func main() {

	rootCmd.AddCommand(cmdServer(), cmdClient())

	if err := rootCmd.Execute(); err != nil {
		return
	}
	fmt.Println("ready for waiting")
	waitForSignal()
	return
}
