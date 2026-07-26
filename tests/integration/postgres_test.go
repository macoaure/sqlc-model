package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const databaseURLEnv = "SQLC_RICHMODEL_TEST_DATABASE_URL"

type postgresTestDB struct {
	URL    string
	Schema string
}

func requirePostgres(t *testing.T) postgresTestDB {
	t.Helper()

	url := os.Getenv(databaseURLEnv)
	if url == "" {
		t.Skipf("%s is not set", databaseURLEnv)
	}

	name := strings.NewReplacer("-", "_", ":", "_", ".", "_").Replace(t.Name())
	schema := fmt.Sprintf("richmodel_%d_%s", time.Now().UnixNano(), strings.ToLower(name))
	return postgresTestDB{URL: url, Schema: schema}
}

func TestRequirePostgresSkipsWithoutURL(t *testing.T) {
	t.Setenv(databaseURLEnv, "")
	requirePostgres(t)
	t.Fatal("requirePostgres should skip without a database URL")
}

func runUserIntegrationE2E(t *testing.T) {
	t.Helper()

	db := requirePostgres(t)
	dir := filepath.Join("..", "compile", "user-integration")
	cmd := exec.Command("go", "run", "./e2e")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "DATABASE_URL="+db.URL, "RICHMODEL_TEST_SCHEMA="+db.Schema)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("user-integration e2e failed for schema %s: %v\n%s", db.Schema, err, out)
	}
}
