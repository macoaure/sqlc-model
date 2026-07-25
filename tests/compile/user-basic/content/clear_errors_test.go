package content

import (
	"errors"
	"reflect"
	"testing"
)

func TestClearErrorsPreservesValuesAndDirtyState(t *testing.T) {
	user := (&User{}).SetName("Ada").SetEmail("ada@example.com")
	dirtyBefore := user.DirtyFields()
	nameBefore := user.Name()
	emailBefore := user.Email()

	user.setFieldError(UserFieldName, errors.New("name is invalid"))
	user.setFieldError(UserFieldEmail, errors.New("email is invalid"))
	user.ClearErrors()

	if user.HasErrors() {
		t.Fatalf("ClearErrors() should clear validation state, got %v", user.Err())
	}
	if user.Name() != nameBefore || user.Email() != emailBefore {
		t.Fatal("ClearErrors() should not alter field values")
	}
	if !reflect.DeepEqual(user.DirtyFields(), dirtyBefore) {
		t.Fatalf("DirtyFields() = %v, want %v", user.DirtyFields(), dirtyBefore)
	}
}
