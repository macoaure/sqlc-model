package content

import (
	"errors"
	"testing"
)

func TestOperationContractErrorsArePublic(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrDetachedModel, ErrDeletedModel, ErrOperationNotConfigured} {
		if err == nil || !errors.Is(err, err) {
			t.Fatalf("contract error should be non-nil and comparable through errors.Is: %v", err)
		}
	}
}
