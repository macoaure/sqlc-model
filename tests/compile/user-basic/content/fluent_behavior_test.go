package content

import "testing"

func (u *User) Rename(name string) *User {
	return u.SetName(name)
}

func TestFluentCustomMethodsChainWithGeneratedSetters(t *testing.T) {
	user := (&User{}).Rename("Ada").Activate().SetEmail("ada@example.com")

	if user.Name() != "Ada" {
		t.Fatalf("Name() = %q, want Ada", user.Name())
	}
	if !user.Active() {
		t.Fatal("Activate() should set Active")
	}
	if user.Email() != "ada@example.com" {
		t.Fatalf("Email() = %q, want ada@example.com", user.Email())
	}
	if !user.IsDirty(UserFieldName, UserFieldActive, UserFieldEmail) {
		t.Fatal("custom methods should compose with generated dirty tracking")
	}
}
