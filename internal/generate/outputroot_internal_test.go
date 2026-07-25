package generate

import (
	"os"
	"path/filepath"
	"testing"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestOutputRoot_RelativeJoinsWithCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	req := &pb.GenerateRequest{Settings: &pb.Settings{Codegen: &pb.Codegen{Out: "gen"}}}
	got := outputRoot(req)
	want := filepath.Join(cwd, "gen")
	if got != want {
		t.Fatalf("outputRoot() = %q, want %q", got, want)
	}
}

// TestOutputRoot_GetwdFailure covers the os.Getwd() error branch: removing
// the process's current directory out from under it (permitted on Linux)
// makes any subsequent os.Getwd() call fail.
func TestOutputRoot_GetwdFailure(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	t.Cleanup(func() {
		if cerr := os.Chdir(orig); cerr != nil {
			t.Fatalf("failed to restore original working directory: %v", cerr)
		}
	})

	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("os.Mkdir: %v", err)
	}
	if err := os.Chdir(sub); err != nil {
		t.Fatalf("os.Chdir: %v", err)
	}
	if err := os.Remove(sub); err != nil {
		t.Skipf("platform does not allow removing the current directory: %v", err)
	}

	req := &pb.GenerateRequest{Settings: &pb.Settings{Codegen: &pb.Codegen{Out: "gen"}}}
	if got := outputRoot(req); got != "" {
		t.Fatalf("outputRoot() = %q, want empty string when os.Getwd fails", got)
	}
}
