package golden

import (
	"strings"
	"testing"

	"github.com/macoaure/sqlc-model/internal/diagnostics"
	"github.com/macoaure/sqlc-model/internal/generate"
)

func TestGoldenDiagnosticsOutput(t *testing.T) {
	req := userBasicRequest(t)
	for _, q := range req.Queries {
		if q.Name == "CreateUser" {
			q.Cmd = ":exec"
		}
	}

	resp, diags := generate.Generate(req)
	if resp != nil {
		t.Fatalf("expected diagnostic-only failure, got %d files", len(resp.Files))
	}
	text := diagnostics.FormatAll(diagnostics.Sort(diags))
	for _, want := range []string{"error:", "CreateUser", ":exec"} {
		if !strings.Contains(text, want) {
			t.Fatalf("diagnostic output missing %q:\n%s", want, text)
		}
	}
}
