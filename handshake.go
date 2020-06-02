package lurker

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/portmapping/lurker/common"
)

// HandshakeStatusSuccess ...
const HandshakeStatusSuccess HandshakeStatus = 0x01

// HandshakeStatusFailed ...
const HandshakeStatusFailed HandshakeStatus = 0x00

// HandshakeTypePing ...
const HandshakeTypePing HandshakeType = 0x01

// HandshakeTypeConnect ...
const HandshakeTypeConnect HandshakeType = 0x02

// HandshakeTypeAdapter ...
const HandshakeTypeAdapter HandshakeType = 0x03

// HandshakeAuthorization ...
const HandshakeAuthorization HandshakeType = 0x04

// HandshakeRequestTypeProxy ...
const HandshakeRequestTypeProxy RequestType = 0x01

// Version ...
type Version [4]byte

// HandshakeStatus ...
type HandshakeStatus uint8

// HandshakeType ...
type HandshakeType uint8

// RequestType ...
type RequestType int

// HandshakeHead ...
type HandshakeHead struct {
	Type    HandshakeType `json:"type"`
	Tunnel  uint8         `json:"tunnel"`
	Version Version       `json:"version"`
}

// HandshakeResponder ...
type HandshakeResponder interface {
	Pong() error
	Intermediary() error
	Interaction() error
	Other() error
}

// HandshakeRequester ...
type HandshakeRequester interface {
	Ping() error
	Connect() error
	Adapter() error
}

// HandshakeRequest ...
type HandshakeRequest struct {
	Head HandshakeHead `json:"head"`
	Data []byte        `json:"data"`
}

// HandshakeResponse ...
type HandshakeResponse struct {
	//RequestType RequestType     `json:"request_type"`
	Status HandshakeStatus `json:"status"`
	Data   []byte          `json:"data"`
}

// JSON ...
func (r *HandshakeResponse) JSON() []byte {
	marshal, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return marshal
}

// Service ...
type Service struct {
	ID          string        `json:"id"`
	Addr        []common.Addr `json:"common"`
	ISP         net.IP        `json:"isp"`
	Local       net.IP        `json:"local"`
	PortUDP     int           `json:"port_udp"`
	PortTCP     int           `json:"port_tcp"`
	KeepConnect bool          `json:"keep_connect"`
}

// ParseHandshake ...
func ParseHandshake(data []byte) (HandshakeHead, error) {
	var h HandshakeHead
	err := json.Unmarshal(data, &h)
	if err != nil {
		return HandshakeHead{}, err
	}
	return h, nil
}

// DecodeHandshakeRequest ...
func DecodeHandshakeRequest(data []byte, r *HandshakeRequest) (Service, error) {
	err := json.Unmarshal(data, r)
	if err != nil {
		return Service{}, err
	}
	return decodeHandshakeRequestV1(r)
}

func decodeHandshakeRequestV1(request *HandshakeRequest) (Service, error) {
	var s Service
	err := json.Unmarshal(request.Data, &s)
	if err != nil {
		return Service{}, err
	}
	return s, nil
}

// EncodeHandshakeRequest ...
func EncodeHandshakeRequest(service Service) ([]byte, error) {
	return encodeHandshakeRequestV1(&HandshakeRequest{
		Data: service.JSON(),
	})
}
func encodeHandshakeRequestV1(request *HandshakeRequest) ([]byte, error) {
	return json.Marshal(request)
}

// EncodeHandshakeResponse ...
func EncodeHandshakeResponse(ver Version, r *HandshakeResponse) ([]byte, error) {
	switch ver {
	default:

	}
	return encodeHandshakeResponseV1(r)
}

func encodeHandshakeResponseV1(r *HandshakeResponse) ([]byte, error) {
	return json.Marshal(r)
}

func decodeHandshakeResponse(data []byte) (*HandshakeResponse, error) {
	var resp HandshakeResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// JSON ...
func (h HandshakeHead) JSON() []byte {
	marshal, err := json.Marshal(h)
	if err != nil {
		return nil
	}
	return marshal
}

// ParseHandshakeHead ...
func ParseHandshakeHead(b []byte) (HandshakeHead, error) {
	var h HandshakeHead
	if len(b) < 8 {
		return h, fmt.Errorf("wrong byte size")
	}
	h.Type = HandshakeType(b[0])
	h.Tunnel = b[1]
	h.Version[0] = b[4]
	h.Version[1] = b[5]
	h.Version[2] = b[6]
	h.Version[3] = b[7]
	return h, nil
}

// Bytes ...
func (h HandshakeHead) Bytes() []byte {
	var dummy uint8 = 0
	b := []byte{
		uint8(h.Type),
		h.Tunnel,
		dummy,
		dummy,
	}
	b = append(b, h.Version[:]...)
	return b
}

// Run ...
func (h *HandshakeHead) Run(able HandshakeResponder) error {
	switch h.Type {
	case HandshakeTypePing:
		return able.Pong()
	case HandshakeTypeConnect:
		return able.Interaction()
	case HandshakeTypeAdapter:
		return able.Intermediary()
	}
	return able.Other()
}
