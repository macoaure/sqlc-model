package content

import (
	"errors"
	"testing"
)

func TestFieldErrorCorrectionKeepsUnrelatedErrors(t *testing.T) {
	nameErr := errors.New("name is required")
	emailErr := errors.New("email is required")
	user := &User{}

	user.setFieldError(UserFieldName, nameErr)
	user.setFieldError(UserFieldEmail, emailErr)
	user.setFieldError(UserFieldName, errors.New("name is still required"))

	if errors.Is(user.FieldError(UserFieldName), nameErr) {
		t.Fatal("same-field errors should be replaced")
	}
	if !errors.Is(user.FieldError(UserFieldEmail), emailErr) {
		t.Fatal("unrelated field error should remain")
	}

	user.SetName("Ada")
	if err := user.FieldError(UserFieldName); err != nil {
		t.Fatalf("valid setter should clear field error, got %v", err)
	}
	if !errors.Is(user.FieldError(UserFieldEmail), emailErr) {
		t.Fatal("clearing name should not clear email")
	}
}
