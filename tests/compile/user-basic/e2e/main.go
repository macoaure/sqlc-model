// Command e2e exercises quickstart.md's full walkthrough against a real
// PostgreSQL instance, using only the generated model API — no
// hand-written SQL, struct-scanning, or manual dirty-flag code (T050).
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/macoaure/sqlc-gen-richmodel/tests/compile/user-basic/content"
)

func must(cond bool, msg string) {
	if !cond {
		log.Fatalf("FAIL: %s", msg)
	}
}

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting: %v", err)
	}
	defer pool.Close()

	sess := content.New(pool)

	user := sess.Users.New().
		SetName("Marcos Aurelio").
		SetEmail("marcos@example.com").
		SetActive(true)

	must(user.IsNew(), "new model should report IsNew()")
	must(user.IsDirty(), "new model with set fields should be dirty")

	if err := user.Save(ctx); err != nil {
		log.Fatalf("Save (insert) failed: %v", err)
	}
	fmt.Println("inserted user, id =", user.ID())

	must(user.Exists(), "Exists() should be true after insert")
	must(!user.IsDirty(), "model should be clean after insert (snapshot synchronized)")

	found, err := sess.Users.Find(ctx, user.ID())
	if err != nil {
		log.Fatalf("Find failed: %v", err)
	}
	must(found.Name() == "Marcos Aurelio", "found.Name() should round-trip")
	must(found.Exists() && !found.IsDeleted(), "found model should be persisted, not deleted")

	found.SetName("Marcos A.")
	must(found.IsDirty(), "model should be dirty after SetName changes value")
	if err := found.Save(ctx); err != nil {
		log.Fatalf("Save (update) failed: %v", err)
	}
	must(!found.IsDirty(), "model should be clean after update")
	must(found.Name() == "Marcos A.", "update should persist new name")

	if err := found.Refresh(ctx); err != nil {
		log.Fatalf("Refresh failed: %v", err)
	}
	must(found.Name() == "Marcos A.", "refresh should re-hydrate the same persisted name")

	if err := found.Delete(ctx); err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	must(found.IsDeleted(), "IsDeleted() should be true after Delete")

	if err := found.Delete(ctx); err != nil {
		log.Fatalf("second Delete should be idempotent, got error: %v", err)
	}

	if err := found.Save(ctx); err == nil {
		log.Fatal("Save on a deleted model should return an error")
	} else {
		fmt.Println("Save on deleted model correctly returned:", err)
	}

	fmt.Println("ALL CHECKS PASSED")
}
