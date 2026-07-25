package golden

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/generate"
)

func TestLifecycleGeneratedAPISurface(t *testing.T) {
	resp, diags := generate.Generate(userBasicRequest(t))
	if resp == nil {
		t.Fatalf("generation failed: %v", diags)
	}

	model := generatedFile(t, resp.Files, "content/user_gen.go")
	for _, want := range []string{
		"func (u *User) IsPersisted() bool",
		"func (u *User) HasChanges() bool",
		"func (u *User) Detach() *User",
		"return ErrModelDeleted",
		"return ErrModelDetached",
		"return ErrModelNotPersisted",
		"u.current, u.original = out, out",
		"u.original = u.current",
	} {
		if !strings.Contains(model, want) {
			t.Fatalf("expected generated model to contain %q:\n%s", want, model)
		}
	}

	session := generatedFile(t, resp.Files, "content/session_gen.go")
	for _, want := range []string{
		"ErrModelDeleted = errors.New(\"richmodel: model is deleted\")",
		"ErrModelDetached = errors.New(\"richmodel: model is not attached to a session\")",
		"ErrModelNotPersisted = errors.New(\"richmodel: model is not persisted\")",
		"ErrDeletedModel = ErrModelDeleted",
		"ErrDetachedModel = ErrModelDetached",
	} {
		if !strings.Contains(session, want) {
			t.Fatalf("expected generated session to contain %q:\n%s", want, session)
		}
	}
}
