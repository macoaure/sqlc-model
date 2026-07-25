package content

import (
	"context"
	"errors"
	"testing"
)

func TestCollectionShortcutAPISignatures(t *testing.T) {
	s := New(nil)
	u := s.Users.New()
	if !u.IsAttached() || !u.IsNew() {
		t.Fatalf("new user should be attached and unpersisted")
	}

	var _ func(*UserCollection, context.Context, string) (*User, error) = (*UserCollection).FindByName
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Fatalf("ErrNotFound should be usable with errors.Is")
	}
}
