package golden

import "testing"

type generatedModelFixture struct {
	Name  string
	Cases []string
}

func generatedModelFixtures() []generatedModelFixture {
	return []generatedModelFixture{
		{Name: "user-basic", Cases: []string{"uuid identifier", "timestamps", "root session"}},
		{Name: "user-query", Cases: []string{":one", ":many", "has-many", "many-to-many"}},
		{Name: "user-relations", Cases: []string{"belongs-to", "has-many", "eager loading"}},
		{Name: "user-value-object", Cases: []string{"custom override", "nullable custom type"}},
	}
}

func TestGeneratedModelFixtureMetadata(t *testing.T) {
	seen := map[string]bool{}
	for _, fixture := range generatedModelFixtures() {
		if fixture.Name == "" {
			t.Fatal("fixture name is required")
		}
		if seen[fixture.Name] {
			t.Fatalf("duplicate fixture name %q", fixture.Name)
		}
		seen[fixture.Name] = true
		if len(fixture.Cases) == 0 {
			t.Fatalf("fixture %s has no coverage cases", fixture.Name)
		}
	}
}
