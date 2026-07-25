package content

import (
	"errors"
	"testing"
)

func TestLifecycleAPISurfaceCompiles(t *testing.T) {
	u := (&User{}).Detach()

	if !u.IsDetached() {
		t.Fatal("detached model should report detached state")
	}
	if u.IsPersisted() {
		t.Fatal("zero model should not report persisted state")
	}
	if u.HasChanges() {
		t.Fatal("zero model should not report changes")
	}
	if !errors.Is(ErrDeletedModel, ErrModelDeleted) {
		t.Fatal("legacy deleted error should wrap lifecycle sentinel")
	}
	if !errors.Is(ErrDetachedModel, ErrModelDetached) {
		t.Fatal("legacy detached error should wrap lifecycle sentinel")
	}
	_ = ErrModelNotPersisted
}

func TestLifecycleOriginalAccessorsCompile(t *testing.T) {
	u := &User{}
	u.SetName("Ada")
	u.SetEmail("ada@example.com")
	u.SetActive(true)

	_ = u.OriginalID()
	_ = u.OriginalName()
	_ = u.OriginalEmail()
	_ = u.OriginalActive()
	_ = u.OriginalCreatedAt()
	_ = u.OriginalUpdatedAt()
}
