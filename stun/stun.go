package stun

import (
	"fmt"
	"github.com/ccding/go-stun/stun"
)

func Support() {
	nat, host, err := stun.NewClient().Discover()
	fmt.Println(nat, host, err)
}
