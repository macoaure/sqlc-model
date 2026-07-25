package content

import "testing"

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
