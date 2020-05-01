package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var tag string

const HandShakeMsg = "我是打洞消息"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("请输入一个客户端标志")
		os.Exit(0)
	}

}

func handleTCP() {
	// 当前进程标记字符串,便于显示
	tag = os.Args[1]
	srcAddr := &net.TCPAddr{IP: net.IPv4zero, Port: 16005} // 注意端口必须固定
	dstAddr := &net.TCPAddr{IP: net.ParseIP("47.101.169.94"), Port: 16004}
	conn, err := net.DialTCP("tcp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	if _, err = conn.Write([]byte("hello, I'm new peer:" + tag)); err != nil {
		log.Panic(err)
	}

	//data := make([]byte, 1024)
	//conn.ReadFrom()
	//n, remoteAddr, err := conn.ReadFromUDP(data)
	//if err != nil {
	//	fmt.Printf("error during read: %s", err)
	//}
	//conn.Close()
	//anotherPeer := parseAddr(string(data[:n]))
	//fmt.Printf("local:%s server:%s another:%sn", srcAddr, remoteAddr, anotherPeer.String())
	// 开始打洞
	//bidirectionalHoleTCP(srcAddr, conn)
}

func handleUDP() {
	// 当前进程标记字符串,便于显示
	tag = os.Args[1]
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 16005} // 注意端口必须固定
	dstAddr := &net.UDPAddr{IP: net.ParseIP("47.101.169.94"), Port: 16004}
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
	anotherPeer := parseAddr(string(data[:n]))
	fmt.Printf("local:%s server:%s another:%sn", srcAddr, remoteAddr, anotherPeer.String())
	// 开始打洞
	bidirectionalHoleUDP(srcAddr, &anotherPeer)
}

func parseAddr(addr string) net.UDPAddr {
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
			log.Printf("error during read: %sn", err)
		} else {
			log.Printf("收到数据:%sn", data[:n])
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
		n, err := conn.Read(data)
		if err != nil {
			log.Printf("error during read: %sn", err)
		} else {
			log.Printf("收到数据:%sn", data[:n])
		}
	}
}
