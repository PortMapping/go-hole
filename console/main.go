package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	ip := "127.0.0.1"
	port := 18080
	if len(os.Args) > 1 {
		ip = os.Args[1]
	}

	if len(os.Args) > 2 {
		parseInt, err := strconv.ParseInt(os.Args[2], 10, 32)
		if err == nil {
			port = int(parseInt)
		}
	}
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// handle error
	}
	_, err = conn.Write([]byte("hello world"))
	if err != nil {
		panic(err)
	}

}
