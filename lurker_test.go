package lurker

import (
	"github.com/portmapping/lurker/common"
	"strings"
	"testing"
)

// TestParseAddr ...
func TestParseAddr(t *testing.T) {
	addr, i := common.ParseAddr("192.168.0.0:1234")
	if strings.Compare("192.168.0.0", addr.String()) != 0 {
		t.Fatal(addr.String(), i)
	}
	if i != 1234 {
		t.Fatal(addr.String(), i)
	}
	addr, i = common.ParseAddr("192.168.0.0")
	if strings.Compare("192.168.0.0", addr.String()) != 0 {
		t.Fatal(addr.String(), i)
	}
}
