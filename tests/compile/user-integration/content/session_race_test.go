package content

import (
	"sync"
	"testing"
)

func TestConcurrentSessionConstruction(t *testing.T) {
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sess := New(nil)
			if sess.Users == nil || sess.Posts == nil || sess.Tags == nil {
				t.Error("session collections not initialized")
			}
		}()
	}
	wg.Wait()
}
