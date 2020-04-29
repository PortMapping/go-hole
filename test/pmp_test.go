package test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackpal/gateway"
	natpmp "github.com/jackpal/go-nat-pmp"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

type callRecord struct {
	// The expected msg argument to call.
	msg    []byte
	result []byte
	err    error
}

type mockNetwork struct {
	// test object, used to report errors.
	t  *testing.T
	cr callRecord
}

func (n *mockNetwork) call(msg []byte, timeout time.Duration) (result []byte, err error) {
	if bytes.Compare(msg, n.cr.msg) != 0 {
		n.t.Errorf("msg=%v, expected %v", msg, n.cr.msg)
	}
	return n.cr.result, n.cr.err
}

type getExternalAddressRecord struct {
	result *natpmp.GetExternalAddressResult
	err    error
	cr     callRecord
}

func getClient() *natpmp.Client {
	gatewayIP, err := gateway.DiscoverGateway()
	if err != nil {
		panic(err)
	}

	return natpmp.NewClient(gatewayIP)
}

func TestPMP(t *testing.T) {
	gatewayIP, err := gateway.DiscoverGateway()
	if err != nil {
		t.Fatal(err)
	}

	client := natpmp.NewClient(gatewayIP)
	response, err := client.GetExternalAddress()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("External IP address: %v\n", response.ExternalIPAddress)
	var localIP net.IP
	// Port comes from the natpmp package
	timeoutCtx, cancel := context.WithTimeout(context.TODO(), 30)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(timeoutCtx, "udp", net.JoinHostPort(gatewayIP.String(), "5351"))
	if err == nil {
		conn.Close()
		localIPAddress, _, err := net.SplitHostPort(conn.LocalAddr().String())
		if err == nil {
			localIP = net.ParseIP(localIPAddress)
		} else {
			t.Fatal("Failed to lookup local IP", err)
		}
	}
	t.Log("local ip", localIP.To4())
}

type callRecorder struct {
	callRecord callRecord
}

func (cr *callRecorder) observeCall(msg []byte, result []byte, err error) {
	cr.callRecord = callRecord{msg, result, err}
}
func TestRecordGetExternalAddress(t *testing.T) {
	c := getClient()
	result, err := c.GetExternalAddress()
	t.Logf("%+v, %+v", result, err)
}

func TestRecordAddPortMapping(t *testing.T) {
	c := getClient()
	mapping, err := c.AddPortMapping("tcp", 10080, 18080, 3600)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", mapping)

	ln, err := net.Listen("tcp", ":10080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	all, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("received:", string(all))
	return
}
