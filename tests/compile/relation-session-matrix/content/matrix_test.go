package content

import "testing"

func TestRelationSessionMatrixCases(t *testing.T) {
	for _, name := range []string{
		"belongs-to",
		"has-one",
		"has-many",
		"many-to-many",
		"root-session",
		"transaction-session",
		"mismatched-session",
	} {
		t.Run(name, func(t *testing.T) {})
	}
}
