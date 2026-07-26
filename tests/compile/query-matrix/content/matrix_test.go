package content

import "testing"

func TestQueryMatrixCases(t *testing.T) {
	for _, name := range []string{
		"zero-parameters",
		"one-parameter",
		"multiple-parameters",
		"table-row-result",
		"custom-row-result",
		"joined-row-result",
		"cmd-one",
		"cmd-many",
		"cmd-exec",
		"cmd-execrows",
	} {
		t.Run(name, func(t *testing.T) {})
	}
}
