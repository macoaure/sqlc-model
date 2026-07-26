package content

import "testing"

func TestTypeMatrixCases(t *testing.T) {
	for _, name := range []string{
		"pointer-nullability",
		"pgtype-wrapper-nullability",
		"nullable-custom-type",
		"text",
		"boolean",
		"numeric",
		"json",
		"jsonb",
		"byte-array",
		"enum",
		"array",
		"timestamp",
	} {
		t.Run(name, func(t *testing.T) {})
	}
}
