package test

import (
	"bytes"
	"fmt"
	"github.com/jackpal/gateway"
	natpmp "github.com/jackpal/go-nat-pmp"
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
	t.Logf("%#v, %#v", result, err)
}

func TestRecordAddPortMapping(t *testing.T) {
	cr := &callRecorder{}
	c := getClient()
	result, err := c.AddPortMapping("tcp", 123, 0, 0)
	t.Logf("%#v, %#v, %#v", result, err, cr.callRecord)
}
func TestGetExternalAddress(t *testing.T) {
	dummyError := fmt.Errorf("dummy error")
	testCases := []getExternalAddressRecord{
		{
			nil,
			dummyError,
			callRecord{[]uint8{0x0, 0x0}, nil, dummyError},
		},
		{
			&natpmp.GetExternalAddressResult{SecondsSinceStartOfEpoc: 0x13f24f, ExternalIPAddress: [4]uint8{0x49, 0x8c, 0x36, 0x9a}},
			nil,
			callRecord{[]uint8{0x0, 0x0}, []uint8{0x0, 0x80, 0x0, 0x0, 0x0, 0x13, 0xf2, 0x4f, 0x49, 0x8c, 0x36, 0x9a}, nil},
		},
	}
	for i, testCase := range testCases {
		t.Logf("case %d", i)
		c := getClient()
		result, err := c.GetExternalAddress()
		if err != nil {
			if err != testCase.err {
				t.Error(err)
			}
		} else {
			if result.SecondsSinceStartOfEpoc != testCase.result.SecondsSinceStartOfEpoc {
				t.Errorf("result.SecondsSinceStartOfEpoc=%d != %d", result.SecondsSinceStartOfEpoc, testCase.result.SecondsSinceStartOfEpoc)
			}
			if bytes.Compare(result.ExternalIPAddress[:], testCase.result.ExternalIPAddress[:]) != 0 {
				t.Errorf("result.ExternalIPAddress=%v != %v", result.ExternalIPAddress, testCase.result.ExternalIPAddress)
			}
		}
	}
}

type addPortMappingRecord struct {
	protocol              string
	internalPort          int
	requestedExternalPort int
	lifetime              int
	result                *natpmp.AddPortMappingResult
	err                   error
	cr                    callRecord
}

func TestAddPortMapping(t *testing.T) {
	dummyError := fmt.Errorf("dummy error")
	testCases := []addPortMappingRecord{
		// Propagate error
		{
			protocol: "udp", internalPort: 123, requestedExternalPort: 456, lifetime: 1200,
			err: dummyError,
			cr:  callRecord{[]uint8{0x0, 0x1, 0x0, 0x0, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0}, nil, dummyError},
		},
		// Add UDP
		{
			protocol: "udp", internalPort: 123, requestedExternalPort: 456, lifetime: 1200,
			result: &natpmp.AddPortMappingResult{SecondsSinceStartOfEpoc: 0x13feff, InternalPort: 0x7b, MappedExternalPort: 0x1c8, PortMappingLifetimeInSeconds: 0x4b0},
			cr: callRecord{
				msg:    []uint8{0x0, 0x1, 0x0, 0x0, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
				result: []uint8{0x0, 0x81, 0x0, 0x0, 0x0, 0x13, 0xfe, 0xff, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
			},
		},
		// Add TCP
		{
			protocol: "tcp", internalPort: 123, requestedExternalPort: 456, lifetime: 1200,
			result: &natpmp.AddPortMappingResult{SecondsSinceStartOfEpoc: 0x140321, InternalPort: 0x7b, MappedExternalPort: 0x1c8, PortMappingLifetimeInSeconds: 0x4b0},
			cr: callRecord{
				msg:    []uint8{0x0, 0x2, 0x0, 0x0, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
				result: []uint8{0x0, 0x82, 0x0, 0x0, 0x0, 0x14, 0x3, 0x21, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
			},
		},
		// Remove UDP
		{
			protocol: "udp", internalPort: 123,
			result: &natpmp.AddPortMappingResult{SecondsSinceStartOfEpoc: 0x1403d5, InternalPort: 0x7b},
			cr: callRecord{
				msg:    []uint8{0x0, 0x1, 0x0, 0x0, 0x0, 0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				result: []uint8{0x0, 0x81, 0x0, 0x0, 0x0, 0x14, 0x3, 0xd5, 0x0, 0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
		},
		// Remove TCP
		{
			protocol: "tcp", internalPort: 123,
			result: &natpmp.AddPortMappingResult{SecondsSinceStartOfEpoc: 0x140496, InternalPort: 0x7b},
			cr: callRecord{
				msg:    []uint8{0x0, 0x2, 0x0, 0x0, 0x0, 0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				result: []uint8{0x0, 0x82, 0x0, 0x0, 0x0, 0x14, 0x4, 0x96, 0x0, 0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
		},
	}

	for i, testCase := range testCases {
		t.Logf("case %d", i)
		c := getClient()
		result, err := c.AddPortMapping(testCase.protocol, testCase.internalPort, testCase.requestedExternalPort, testCase.lifetime)
		if err != nil || testCase.err != nil {
			if err != testCase.err && fmt.Sprintf("%v", err) != fmt.Sprintf("%v", testCase.err) {
				t.Errorf("err=%v != %v", err, testCase.err)
			}
		} else {
			if result.SecondsSinceStartOfEpoc != testCase.result.SecondsSinceStartOfEpoc {
				t.Errorf("result.SecondsSinceStartOfEpoc=%d != %d", result.SecondsSinceStartOfEpoc, testCase.result.SecondsSinceStartOfEpoc)
			}
			if result.InternalPort != testCase.result.InternalPort {
				t.Errorf("result.InternalPort=%d != %d", result.InternalPort, testCase.result.InternalPort)
			}
			if result.MappedExternalPort != testCase.result.MappedExternalPort {
				t.Errorf("result.InternalPort=%d != %d", result.MappedExternalPort, testCase.result.MappedExternalPort)
			}
			if result.PortMappingLifetimeInSeconds != testCase.result.PortMappingLifetimeInSeconds {
				t.Errorf("result.InternalPort=%d != %d", result.PortMappingLifetimeInSeconds, testCase.result.PortMappingLifetimeInSeconds)
			}
		}
	}
}

func TestProtocolChecks(t *testing.T) {
	testCases := []addPortMappingRecord{
		// Unexpected result size.
		{
			protocol: "tcp", internalPort: 123, requestedExternalPort: 456, lifetime: 1200,
			err: fmt.Errorf("unexpected result size %d, expected %d", 1, 16),
			cr: callRecord{
				[]uint8{0x0, 0x2, 0x0, 0x0, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
				[]uint8{0x0},
				nil,
			},
		},
		//  Unknown protocol version.
		{
			protocol: "tcp", internalPort: 123, requestedExternalPort: 456, lifetime: 1200,
			err: fmt.Errorf("unknown protocol version %d", 1),
			cr: callRecord{
				[]uint8{0x0, 0x2, 0x0, 0x0, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
				[]uint8{0x1, 0x82, 0x0, 0x0, 0x0, 0x14, 0x4, 0x96, 0x0, 0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				nil,
			},
		},
		// Unexpected opcode.
		{
			protocol: "tcp", internalPort: 123, requestedExternalPort: 456, lifetime: 1200,
			err: fmt.Errorf("unexpected opcode %d. Expected %d", 0x88, 0x82),
			cr: callRecord{
				[]uint8{0x0, 0x2, 0x0, 0x0, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
				[]uint8{0x0, 0x88, 0x0, 0x0, 0x0, 0x14, 0x4, 0x96, 0x0, 0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				nil,
			},
		},
		// Non-success result code.
		{
			protocol: "tcp", internalPort: 123, requestedExternalPort: 456, lifetime: 1200,
			err: fmt.Errorf("non-zero result code %d", 17),
			cr: callRecord{
				[]uint8{0x0, 0x2, 0x0, 0x0, 0x0, 0x7b, 0x1, 0xc8, 0x0, 0x0, 0x4, 0xb0},
				[]uint8{0x0, 0x82, 0x0, 0x11, 0x0, 0x14, 0x4, 0x96, 0x0, 0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				nil,
			},
		},
	}
	for i, testCase := range testCases {
		t.Logf("case %d", i)
		c := getClient()
		result, err := c.AddPortMapping(testCase.protocol, testCase.internalPort, testCase.requestedExternalPort, testCase.lifetime)
		if err != testCase.err && fmt.Sprintf("%v", err) != fmt.Sprintf("%v", testCase.err) {
			t.Errorf("err=%v != %v", err, testCase.err)
		}
		if result != nil {
			t.Errorf("result=%v != nil", result)
		}
	}
}
