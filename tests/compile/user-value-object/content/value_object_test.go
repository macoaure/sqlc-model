package content

import "testing"

// Compatibility fixture inventory:
// UUID and custom override coverage lives in this module; scalar, temporal,
// boolean, text, numeric, JSON/JSONB, byte array, array, enum, nullable, and
// renamed-field coverage is tracked by docs/content/reference/compatibility.md
// and the root unit compatibility documentation check.
func TestUserEmailUsesValueObject(t *testing.T) {
	email, err := NewEmail("ada@example.com")
	if err != nil {
		t.Fatal(err)
	}

	user := (&User{}).SetEmail(email)
	if user.Email() != email {
		t.Fatalf("Email() = %+v, want %+v", user.Email(), email)
	}
	if user.Email().String() != "ada@example.com" {
		t.Fatalf("Email().String() = %q", user.Email().String())
	}
}
