package content

import (
	"errors"
	"testing"
)

func TestRelationErrorIdentities(t *testing.T) {
	for _, err := range []error{
		ErrSessionMismatch,
		ErrUnsavedRelatedModel,
		ErrLazyLoadingPrevented,
	} {
		if !errors.Is(err, err) {
			t.Fatalf("%v should be usable with errors.Is", err)
		}
	}
	if !errors.Is(ErrUnsavedRelated, ErrUnsavedRelatedModel) {
		t.Fatal("legacy ErrUnsavedRelated should match ErrUnsavedRelatedModel")
	}
}
