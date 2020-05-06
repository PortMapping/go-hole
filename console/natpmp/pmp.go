package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/libp2p/go-nat"
)

func main() {
	port := ":5001"

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	doSub := true
	nat, err := nat.DiscoverGateway()
	if err != nil {
		doSub = false
	}

	if doSub {
		log.Printf("nat type: %s", nat.Type())
		daddr, err := nat.GetDeviceAddress()
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		log.Printf("device address: %s", daddr)

		iaddr, err := nat.GetInternalAddress()
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		log.Printf("internal address: %s", iaddr)

		eaddr, err := nat.GetExternalAddress()
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		log.Printf("external address: %s", eaddr)

		eport, err := nat.AddPortMapping("tcp", 16005, "http", 60)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		log.Printf("test-page: http://%s:%d/", eaddr, eport)

		go func() {
			for {
				time.Sleep(30 * time.Second)

				_, err = nat.AddPortMapping("tcp", 16005, "http", 60)
				if err != nil {
					log.Fatalf("error: %s", err)
				}
			}
		}()
		defer nat.DeletePortMapping("tcp", 16005)

		http.ListenAndServe(port, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(200)
			fmt.Fprintf(rw, "Hello there!\n")
			fmt.Fprintf(rw, "nat type: %s\n", nat.Type())
			fmt.Fprintf(rw, "device address: %s\n", daddr)
			fmt.Fprintf(rw, "internal address: %s\n", iaddr)
			fmt.Fprintf(rw, "external address: %s\n", eaddr)
			fmt.Fprintf(rw, "test-page: http://%s:%d/\n", eaddr, eport)
		}))
	} else {
		log.Println("handling on port", port)
		if err := http.ListenAndServe(port, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(200)
			fmt.Fprintf(rw, "Hello there!\n")
		})); err != nil {
			panic(err)
		}
	}
}
