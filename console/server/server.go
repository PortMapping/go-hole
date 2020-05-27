package main //lurker.go
import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	nw := "udp"
	if len(os.Args) > 1 {
		nw = os.Args[1]
	}
	if strings.Compare(nw, "tcp") == 0 {
		handleTCP()
		return
	}
	handleUDP()
}

func handleTCP() {
	fmt.Println("tcp lurker start")
	//listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: 16004})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	listener, err := net.Listen("tcp", ":16004")
	if err != nil {
		fmt.Println(err)
		return
	}
	data := make([]byte, 1024)
	//peers := make([]net.TCPAddr, 0, 2)
	for {
		acceptTCP, err := listener.Accept()
		fmt.Println("accept listen")
		if err != nil {
			fmt.Println(err)
			return
		}

		go func() {
			go func() {
				for {
					n, err := acceptTCP.Read(data)
					if err != nil {
						fmt.Println(err)
						return
					}
					log.Printf("<%s> %s\n", acceptTCP.RemoteAddr().String(), string(data[:n]))
				}
			}()
			//acceptTCP.Close()
			//common, err := net.ResolveTCPAddr("tcp", acceptTCP.RemoteAddr().String())
			//if err != nil {
			//	fmt.Println(err)
			//	return
			//}
			//acceptTCP.Close()
			go func() {
				dial, err := net.Dial("tcp", acceptTCP.RemoteAddr().String())
				if err != nil {
					fmt.Println(err)
					return
				}
				defer dial.Close()
				for {
					if _, err := dial.Write([]byte("test connect")); err != nil {
						return
					}
					time.Sleep(3 * time.Second)
				}
			}()
			//peers = append(peers, *common)
			//if len(peers) == 2 {
			//	log.Printf("进行UDP打洞,建立 %s <--> %s 的连接\n", peers[0].String(), peers[1].String())
			//	time.Sleep(time.Second * 8)
			//	return
			//}
		}()
	}

}

func handleUDP() {
	fmt.Println("udp lurker start")
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 16004})
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Local Addr: <%s> \n", listener.LocalAddr().String())

	peers := make([]net.UDPAddr, 0, 2)
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		log.Printf("<%s> %s\n", remoteAddr.String(), data[:n])
		peers = append(peers, *remoteAddr)
		if len(peers) == 2 {
			log.Printf("进行UDP打洞,建立 %s <--> %s 的连接\n", peers[0].String(), peers[1].String())
			listener.Write([]byte(peers[1].String()))
			listener.Write([]byte(peers[0].String()))
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
