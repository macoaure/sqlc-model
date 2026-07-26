package content

import (
	"context"
	"errors"
	"testing"
)

func TestDetachedLazyLoadReturnsLifecycleError(t *testing.T) {
	_, err := (&User{}).Detach().Posts().Get(context.Background())
	if !errors.Is(err, ErrModelDetached) {
		t.Fatalf("Posts().Get() error = %v, want ErrModelDetached", err)
	}
}
