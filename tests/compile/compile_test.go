package compile

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFixtureModulesCompile(t *testing.T) {
	root := "."
	matches, err := filepath.Glob(filepath.Join(root, "*", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("no compile fixture modules found")
	}

	for _, mod := range matches {
		dir := filepath.Dir(mod)
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command("go", "test", "./...")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOWORK=off")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("compile fixture %s failed: %v\n%s", name, err, out)
			}
		})
	}
}
