package content

import (
	"sync"
	"testing"
)

func TestConcurrentIndependentModelMutation(t *testing.T) {
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user := New(nil).Users.New()
			user.SetName("race")
			if user.Name() != "race" {
				t.Error("unexpected user name")
			}
		}()
	}
	wg.Wait()
}
