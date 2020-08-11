package lurker

import (
	"fmt"
	"testing"
)

func init() {
	list := []SupportType{
		SupportTypePublicTCP,
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
	s := SupportTypePublicTCP
	s.Add(SupportTypePrivateUDP)
	fmt.Println("support", s)
	//output:33
}

// TestSupportType_Del ...
func TestSupportType_Del(t *testing.T) {
	s := SupportTypePublicTCP
	s.Add(SupportTypePrivateUDP)
	s.Del(SupportTypePublicTCP)
	fmt.Println("support", s)
	//output:32
}
