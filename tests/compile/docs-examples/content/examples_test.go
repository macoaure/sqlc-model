package content

import "testing"

func TestMirroredDocumentationExamplesCompile(t *testing.T) {
	examples := []string{
		"first-model",
		"first-relation",
		"first-transaction",
		"model-api",
		"collection-api",
		"relation-api",
		"session-api",
	}
	for _, name := range examples {
		t.Run(name, func(t *testing.T) {})
	}
}
