package stun

import (
	"fmt"
	"github.com/ccding/go-stun/stun"
)

// Support ...
func Support() {
	cli := stun.NewClient()
	nat, host, err := cli.Discover()
	fmt.Println(nat, host, err)
}
