package codegen

import (
	"encoding/json"
	"testing"

	"github.com/macoaure/sqlc-model/internal/mapping"
)

// TestGoLiteral covers goLiteral's branches directly. Some are unreachable
// from the public config/generate entry points: an empty raw value never
// reaches goLiteral in practice (callers only invoke it when a default/
// value was actually configured — see relationParamExpr and RenderRelation's
// scope loop), and relation.ValidateScopes already rejects malformed JSON
// and unsupported value shapes before plan/codegen ever runs (see
// tests/unit/relation_scope_test.go's TestRelation_InvalidScopeJSONFails
// and TestRelation_ScopeValueTypeMismatchFails). Covered directly here
// since goLiteral is unexported.
func TestGoLiteral(t *testing.T) {
	cases := []struct {
		name    string
		raw     json.RawMessage
		gt      mapping.GoType
		want    string
		wantErr bool
	}{
		{name: "empty", raw: nil, want: "nil"},
		{name: "invalid JSON", raw: json.RawMessage(`{not valid`), wantErr: true},
		{name: "null", raw: json.RawMessage(`null`), want: "nil"},
		{name: "bool", raw: json.RawMessage(`true`), want: "true"},
		{name: "int", raw: json.RawMessage(`5`), gt: mapping.GoType{Expr: "int32"}, want: "5"},
		{name: "float", raw: json.RawMessage(`1.5`), gt: mapping.GoType{Expr: "float64"}, want: "1.5"},
		{name: "unmapped numeric type", raw: json.RawMessage(`5`), gt: mapping.GoType{Expr: "pgtype.Numeric"}, want: "5"},
		{name: "string", raw: json.RawMessage(`"hi"`), want: `"hi"`},
		{name: "unsupported array", raw: json.RawMessage(`[1,2,3]`), wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := goLiteral(c.raw, c.gt)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
