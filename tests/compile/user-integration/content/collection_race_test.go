package content

import (
	"sync"
	"testing"
)

func TestConcurrentCollectionNewUsesIndependentModels(t *testing.T) {
	sess := New(nil)
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user := sess.Users.New().SetName("race")
			if user.Name() != "race" {
				t.Error("unexpected user name")
			}
		}()
	}
	wg.Wait()
}
