package lurker

import (
	"encoding/json"
	"net"
)

// HandshakeStatus ...
type HandshakeStatus int

// Version ...
type Version string

// HandshakeRequest ...
type HandshakeRequest struct {
	ProtocolVersion Version `json:"protocol_version"`
	Data            []byte  `json:"data"`
}

// HandshakeResponseV1 ...
type HandshakeResponseV1 struct {
	Status HandshakeStatus `json:"status"`
	Data   []byte          `json:"data"`
}

// Service ...
type Service struct {
	ID          string `json:"id"`
	Addr        []Addr `json:"addr"`
	ISP         net.IP `json:"isp"`
	Local       net.IP `json:"local"`
	PortUDP     int    `json:"port_udp"`
	PortHole    int    `json:"port_hole"`
	PortTCP     int    `json:"port_tcp"`
	KeepConnect bool   `json:"keep_connect"`
	RequestType int    `json:"request_type"`
}

// DecodeHandshakeRequest ...
func DecodeHandshakeRequest(data []byte) (Service, error) {
	var r HandshakeRequest
	err := json.Unmarshal(data, &r)
	if err != nil {
		return Service{}, err
	}
	return decodeHandshakeRequestV1(&r)
}

func decodeHandshakeRequestV1(request *HandshakeRequest) (Service, error) {
	var s Service
	err := json.Unmarshal(request.Data, &s)
	if err != nil {
		return Service{}, err
	}
	return s, nil
}
