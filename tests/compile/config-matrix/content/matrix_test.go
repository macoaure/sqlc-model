package content

import "testing"

func TestConfigMatrixCases(t *testing.T) {
	for _, name := range []string{"rename", "override", "alias"} {
		t.Run(name, func(t *testing.T) {})
	}
}
