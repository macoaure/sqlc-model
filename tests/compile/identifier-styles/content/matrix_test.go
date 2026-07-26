package content

import "testing"

func TestIdentifierStyleCases(t *testing.T) {
	for _, name := range []string{"uuid", "serial-integer", "bigint", "application-generated"} {
		t.Run(name, func(t *testing.T) {})
	}
}
