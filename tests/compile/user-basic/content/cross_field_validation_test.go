package content

import (
	"errors"
	"testing"
)

var errReservedIdentity = errors.New("reserved identity")

func (u *User) validateUser() error {
	if u.Name() == "Reserved" && u.Email() == "reserved@example.com" {
		return errReservedIdentity
	}
	return nil
}

func TestCrossFieldValidationIsOverallError(t *testing.T) {
	user := (&User{}).SetName("Reserved").SetEmail("reserved@example.com")

	for name, err := range map[string]error{
		"Validate": user.Validate(),
		"Err":      user.Err(),
	} {
		if !errors.Is(err, errReservedIdentity) {
			t.Fatalf("%s() error = %v, want cross-field error", name, err)
		}
	}
	if !user.HasErrors() {
		t.Fatal("HasErrors() should include cross-field validation")
	}
	if err := user.FieldError(UserFieldName); err != nil {
		t.Fatalf("cross-field validation should not set field error, got %v", err)
	}
}
