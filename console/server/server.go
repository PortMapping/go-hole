package main //server.go
import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"
)

func main() {

}

func handleTCP() {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: 16004})
	if err != nil {
		fmt.Println(err)
		return
	}

	acceptTCP, err := listener.AcceptTCP()
	if err != nil {
		return
	}
	//data := make([]byte, 1024)
	peers := make([]net.TCPAddr, 0, 2)
	for {
		all, err := ioutil.ReadAll(acceptTCP)
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Printf("<%s> %sn", acceptTCP.RemoteAddr().String(), all[:])
		addr, err := net.ResolveTCPAddr("tcp", acceptTCP.RemoteAddr().String())
		if err != nil {
			fmt.Println(err)
			return
		}
		peers = append(peers, *addr)
		if len(peers) == 2 {
			log.Printf("进行UDP打洞,建立 %s <--> %s 的连接n", peers[0].String(), peers[1].String())
			time.Sleep(time.Second * 8)
			log.Println("中转服务器退出,仍不影响peers间通信")
			return
		}
	}
}

func handleUDP() {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 16004})
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Local Addr: <%s> \nn", listener.LocalAddr().String())
	log.Printf("Remote Addr:<%s> \n", listener.RemoteAddr().String())
	peers := make([]net.UDPAddr, 0, 2)
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		log.Printf("<%s> %sn", remoteAddr.String(), data[:n])
		peers = append(peers, *remoteAddr)
		if len(peers) == 2 {
			log.Printf("进行UDP打洞,建立 %s <--> %s 的连接n", peers[0].String(), peers[1].String())
			_, err := listener.WriteToUDP([]byte(peers[1].String()), &peers[0])
			if err != nil {
				return
			}
			_, err = listener.WriteToUDP([]byte(peers[0].String()), &peers[1])
			if err != nil {
				return
			}
			time.Sleep(time.Second * 8)
			log.Println("中转服务器退出,仍不影响peers间通信")
			return
		}
	}
}
