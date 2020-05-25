package nat

import (
	"sync"
	"testing"
)

// TestFromLocal ...
func TestFromLocal(t *testing.T) {

	port := 2000
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			nat := defaultNAT()
			_, err := nat.AddPortMapping("tcp", 2000, "testport", 0)
			if err != nil {
				t.Log(err)
			}
		}(port + i)
	}
	wg.Wait()
}
