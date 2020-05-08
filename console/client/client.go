package main

import (
	"fmt"
	"github.com/PortMapping/go-hole/nat"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/portmapping/go-reuse"
)

var tag string

// HandShakeMsg ...
const HandShakeMsg = "我是打洞消息"

func main() {
	h := "udp"
	if len(os.Args) > 1 {
		h = os.Args[1]
	}

	if len(os.Args) < 3 {
		fmt.Println("请输入一个客户端标志")
		os.Exit(0)
	}

	if strings.Compare("udp", h) == 0 {
		handleUDP()
		return
	} else if strings.Compare("m", h) == 0 {
		manualUDP(os.Args[2], "manual")
		return
	} else if strings.Compare("tcp", h) == 0 {
		handleTCP(os.Args[2])
		return
	}
	reuseHandle()

}

func handleTCP(ip string) {
	// 当前进程标记字符串,便于显示
	tag = "local name"
	if len(os.Args) > 3 {
		tag = os.Args[3]
	}
	srcAddr := &net.TCPAddr{IP: net.IPv4zero, Port: 16005} // 注意端口必须固定
	//dstAddr := &net.TCPAddr{IP: net.ParseIP("47.96.140.215"), Port: 16004}

	//conn, err := net.Dial("tcp", "47.101.169.94:16004")
	addr := parseTCPAddr(ip)
	conn, err := net.DialTCP("tcp", srcAddr, &addr)
	//conn.err:=net.Dial("tcp",)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		//fmt.Println("from:", conn.RemoteAddr().String())
		if _, err = conn.Write([]byte("hello, i'm " + tag)); err != nil {
			log.Panic(err)
		}
		time.Sleep(time.Second)
	}
	//fmt.Println("local:", conn.LocalAddr().Network())
	//n, err := nat.DiscoverGateway()
	//if err != nil {
	//	log.Panic(err)
	//	return
	//}

	//go func() {
	//	for {
	//		time.Sleep(30 * time.Second)
	//		_, err = n.AddPortMapping("tcp", 16005, "http", 60)
	//		if err != nil {
	//			log.Fatalf("error: %s", err)
	//		}
	//	}
	//}()
	//defer n.DeletePortMapping("tcp", 16005)

	//data := make([]byte, 1024)
	//conn.ReadFrom()
	//n, remoteAddr, err := conn.ReadFromUDP(data)
	//if err != nil {
	//	fmt.Printf("error during read: %s", err)
	//}
	//conn.Close()
	//anotherPeer := parseUDPAddr(string(data[:n]))
	//fmt.Printf("local:%s lurker:%s another:%sn", srcAddr, remoteAddr, anotherPeer.String())
	// 开始打洞
	//bidirectionalHoleTCP(srcAddr, conn)
}

func handleUDP() {
	// 当前进程标记字符串,便于显示
	tag = os.Args[2]
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 16005} // 注意端口必须固定
	dstAddr := &net.UDPAddr{IP: net.ParseIP("47.96.140.215"), Port: 16004}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	if _, err = conn.Write([]byte("hello, I'm new peer:" + tag)); err != nil {
		log.Panic(err)
	}
	data := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Printf("error during read: %s", err)
	}
	conn.Close()

	anotherPeer := parseUDPAddr(string(data[:n]))
	fmt.Printf("local:%s lurker:%s another:%s\n", srcAddr, remoteAddr, anotherPeer.String())
	// 开始打洞
	bidirectionalHoleUDP(srcAddr, &anotherPeer)
}

func parseTCPAddr(addr string) net.TCPAddr {
	t := strings.Split(addr, ":")
	port, _ := strconv.Atoi(t[1])
	return net.TCPAddr{
		IP:   net.ParseIP(t[0]),
		Port: port,
	}
}
func parseUDPAddr(addr string) net.UDPAddr {
	t := strings.Split(addr, ":")
	port, _ := strconv.Atoi(t[1])
	return net.UDPAddr{
		IP:   net.ParseIP(t[0]),
		Port: port,
	}
}

func bidirectionalHoleUDP(srcAddr *net.UDPAddr, anotherAddr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", srcAddr, anotherAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	// 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息
	if _, err = conn.Write([]byte(HandShakeMsg)); err != nil {
		log.Println("send handshake:", err)
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if _, err = conn.Write([]byte("from [" + tag + "]")); err != nil {
				log.Println("send msg fail", err)
			}
		}
	}()
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Printf("error during read: %s\n", err)
		} else {
			log.Printf("收到数据:%s\n", data[:n])
		}
	}
}

func bidirectionalHoleTCP(srcAddr *net.TCPAddr, anotherAddr *net.TCPAddr) {
	conn, err := net.DialTCP("tcp", srcAddr, anotherAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	// 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息
	if _, err = conn.Write([]byte(HandShakeMsg)); err != nil {
		log.Println("send handshake:", err)
	}
	//go func() {
	for {
		time.Sleep(10 * time.Second)
		if _, err = conn.Write([]byte("from [" + tag + "]")); err != nil {
			log.Println("send msg fail", err)
		}
	}
	//}()
	//for {
	//	data := make([]byte, 1024)
	//	n, err := conn.Read(data)
	//	if err != nil {
	//		log.Printf("error during read: %sn", err)
	//	} else {
	//		log.Printf("收到数据:%sn", data[:n])
	//	}
	//}
}

func reuseHandle() {
	local, err := nat.NewNATFromLocal(16005)
	if err != nil {
		fmt.Println(err)
		return
	}
	mapping, err := local.Mapping()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer local.StopMapping()
	addr := fmt.Sprintf("0.0.0.0:%d", mapping)
	l1, err := reuse.Listen("tcp", "0.0.0.0:16005")
	if err != nil {
		fmt.Println(err)
		return
	}
	c, err := reuse.Dial("tcp", addr, "47.96.140.215:16004")
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println("local:", c.LocalAddr().String())
	//n, err := nat.DiscoverGateway()
	//if err != nil {
	//	log.Panic(err)
	//	return
	//}
	//
	//go func() {
	//	for {
	//		time.Sleep(30 * time.Second)
	//		_, err = n.AddPortMapping("tcp", 16005, "http", 60)
	//		if err != nil {
	//			log.Fatalf("error: %s", err)
	//		}
	//	}
	//}()
	//defer n.DeletePortMapping("tcp", 16005)
	//fmt.Println(l1, c)
	go func() {
		for {
			if _, err = c.Write([]byte(HandShakeMsg)); err != nil {
				log.Println("send handshake:", err)
				return
			}
			time.Sleep(3 * time.Second)
		}
	}()
	//data := make([]byte, 1024)
	//read, err := c.Read(data)
	//if err != nil {
	//	log.Println("read data:", err)
	//	return
	//}
	//addr := parseTCPAddr(string(data[:read]))
	fmt.Println("info sent")
	//c.Close()
	for {
		accept, err := l1.Accept()
		fmt.Println("accept new addr")
		if err != nil {
			fmt.Println(err)
			return
		}
		addr := accept.RemoteAddr().String()
		go func() {
			for {
				data := make([]byte, 1024)
				n, err := accept.Read(data)
				if err != nil {
					log.Printf("error during read: %s\n", err)
					return
				} else {
					log.Printf("from:%s recevied:%s\n", addr, data[:n])
				}
			}
		}()
	}

}

func manualUDP(ip string, tag string) {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 16005} // 注意端口必须固定
	addr := parseUDPAddr(ip)
	bidirectionalHoleUDP(srcAddr, &addr)
}
