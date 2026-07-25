package content

import (
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCompatibilityFixtureUsesSupportedContracts(t *testing.T) {
	user := &User{}
	var id pgtype.UUID
	user.current.ID = id
	if user.ID() != id {
		t.Fatalf("ID() = %v, want %v", user.ID(), id)
	}

	email, err := NewEmail("ada@example.com")
	if err != nil {
		t.Fatal(err)
	}
	user.SetEmail(email)
	if user.Email().String() != "ada@example.com" {
		t.Fatalf("Email().String() = %q", user.Email().String())
	}
}

func TestGeneratedCodeDoesNotGuessNullWrapperFields(t *testing.T) {
	body, err := os.ReadFile("user_store_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), ".Valid") {
		t.Fatalf("generated store must not guess at nullable wrapper internals:\n%s", body)
	}
}
