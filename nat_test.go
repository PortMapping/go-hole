package lurker

import (
	"fmt"
	"testing"
)

func init() {
	list := []SupportType{
		SupportTypePubliccTCP,
		SupportTypePublicUDP,
		SupportTypeProviderTCP,
		SupportTypeProviderUDP,
		SupportTypePrivateTCP,
		SupportTypePrivateUDP,
	}
	for _, supportType := range list {
		fmt.Println("num:", supportType)
	}
}

// TestSupportType_Add ...
func TestSupportType_Add(t *testing.T) {
	s := SupportTypePubliccTCP
	s.Add(SupportTypePrivateUDP)
	fmt.Println("support", s)
	//output:33
}

// TestSupportType_Del ...
func TestSupportType_Del(t *testing.T) {
	s := SupportTypePubliccTCP
	s.Add(SupportTypePrivateUDP)
	s.Del(SupportTypePubliccTCP)
	fmt.Println("support", s)
	//output:32
}
