package proxy

import (
	"errors"
	"io"
	"net"
)

const (
	userPassAuth = uint8(2)
	authVersion  = uint8(1)
	success      = uint8(0)
	failure      = uint8(1)
)

// Authenticate ...
type Authenticate interface {
	NeedAuthenticate() bool
	Auth(conn net.Conn) error
}

// Auth ...
type Auth struct {
	Name string
	Pass string
}

type dummyAuth struct {
}

// NeedAuthenticate ...
func (d dummyAuth) NeedAuthenticate() bool {
	return false
}

// Auth ...
func (d dummyAuth) Auth(conn net.Conn) error {
	return nil
}

// NoAuth ...
func NoAuth() Authenticate {
	return &dummyAuth{}
}

// NeedAuthenticate ...
func (a Auth) NeedAuthenticate() bool {
	return true
}

// Auth ...
func (a Auth) Auth(conn net.Conn) error {
	header := []byte{0, 0}
	if _, err := io.ReadAtLeast(conn, header, 2); err != nil {
		return err
	}
	if header[0] != authVersion {
		return errors.New("the authentication method is not supported")
	}
	nameLen := int(header[1])
	nameLoad := make([]byte, nameLen)
	if _, err := io.ReadAtLeast(conn, nameLoad, nameLen); err != nil {
		return err
	}
	if _, err := conn.Read(header[:1]); err != nil {
		return errors.New("error retrieving password length")
	}
	passLen := int(header[0])
	passLoad := make([]byte, passLen)
	if _, err := io.ReadAtLeast(conn, passLoad, passLen); err != nil {
		return err
	}

	if a.Name == string(nameLoad) && a.Pass == string(passLoad) {
		if _, err := conn.Write([]byte{authVersion, success}); err != nil {
			return err
		}
		return nil
	}
	if _, err := conn.Write([]byte{authVersion, failure}); err != nil {
		return err
	}
	return errors.New("validation failed")

}
