package content

import (
	"errors"
	"testing"
)

func TestValidationErrorReplacementAndInspection(t *testing.T) {
	user := &User{}
	firstNameErr := errors.New("first name error")
	secondNameErr := errors.New("second name error")
	emailErr := errors.New("email error")

	user.setFieldError(UserFieldName, firstNameErr)
	user.setFieldError(UserFieldName, secondNameErr)
	user.setFieldError(UserFieldEmail, emailErr)

	err := user.Validate()
	if errors.Is(err, firstNameErr) {
		t.Fatal("replaced field error should not remain in validation result")
	}
	for _, want := range []error{secondNameErr, emailErr} {
		if !errors.Is(err, want) {
			t.Fatalf("Validate() = %v, want wrapped %v", err, want)
		}
	}

	got := collectValidationErrors(err)
	if got["name"] == nil || !errors.Is(got["name"], secondNameErr) {
		t.Fatalf("name validation error = %v, want %v", got["name"], secondNameErr)
	}
	if got["email"] == nil || !errors.Is(got["email"], emailErr) {
		t.Fatalf("email validation error = %v, want %v", got["email"], emailErr)
	}
}

func collectValidationErrors(err error) map[string]error {
	out := map[string]error{}
	var visit func(error)
	visit = func(err error) {
		if err == nil {
			return
		}
		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			out[validationErr.Field] = validationErr.Err
		}
		if joined, ok := err.(interface{ Unwrap() []error }); ok {
			for _, child := range joined.Unwrap() {
				visit(child)
			}
			return
		}
		if wrapped, ok := err.(interface{ Unwrap() error }); ok {
			visit(wrapped.Unwrap())
		}
	}
	visit(err)
	return out
}
