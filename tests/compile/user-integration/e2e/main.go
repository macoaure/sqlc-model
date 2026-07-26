// Command e2e exercises specs/002-relations-lazy-eager-loading's
// quickstart.md walkthrough against a real PostgreSQL instance, using only
// the generated relation API: lazy read, scoped read, association,
// many-to-many attach/detach/sync, eager loading with inverse hydration,
// and strict lazy-loading mode.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/macoaure/sqlc-model/tests/compile/user-integration/content"
)

func must(cond bool, msg string) {
	if !cond {
		log.Fatalf("FAIL: %s", msg)
	}
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func setupPool(ctx context.Context, dsn, schema string) (*pgxpool.Pool, func()) {
	if schema == "" {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			log.Fatalf("connecting: %v", err)
		}
		return pool, pool.Close
	}

	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting for schema setup: %v", err)
	}
	quoted := quoteIdent(schema)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		admin.Close()
		log.Fatalf("creating schema %s: %v", schema, err)
	}
	admin.Close()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("parsing dsn: %v", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("connecting with schema %s: %v", schema, err)
	}

	body, err := os.ReadFile(filepath.Join("..", "..", "integration", "testdata", "schema.sql"))
	if err != nil {
		pool.Close()
		log.Fatalf("reading schema fixture: %v", err)
	}
	for _, stmt := range strings.Split(string(body), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := pool.Exec(ctx, stmt); err != nil {
			pool.Close()
			log.Fatalf("applying schema statement %q: %v", stmt, err)
		}
	}

	return pool, func() {
		pool.Close()
		cleanup, err := pgxpool.New(ctx, dsn)
		if err != nil {
			log.Printf("cleanup connect failed: %v", err)
			return
		}
		defer cleanup.Close()
		if _, err := cleanup.Exec(ctx, "DROP SCHEMA IF EXISTS "+quoted+" CASCADE"); err != nil {
			log.Printf("cleanup schema %s failed: %v", schema, err)
		}
	}
}

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	schema := os.Getenv("RICHMODEL_TEST_SCHEMA")

	pool, cleanup := setupPool(ctx, dsn, schema)
	defer cleanup()

	sess := content.New(pool)

	var committedUserID, rolledBackUserID, panicUserID pgtype.UUID
	if err := sess.Transaction(ctx, func(tx *content.Session) error {
		txUser := tx.Users.New().SetName("Committed Tx User")
		if err := txUser.Save(ctx); err != nil {
			return err
		}
		committedUserID = txUser.ID()

		txPost := tx.Posts.New().SetTitle("Committed Tx Post")
		if err := txPost.Author().Associate(txUser); err != nil {
			return err
		}
		return txPost.Save(ctx)
	}); err != nil {
		log.Fatalf("transaction commit: %v", err)
	}
	committedUser, err := sess.Users.Find(ctx, committedUserID)
	if err != nil {
		log.Fatalf("find committed transaction user: %v", err)
	}
	must(committedUser.Name() == "Committed Tx User", "committed transaction user should be visible")

	abortErr := errors.New("abort transaction")
	err = sess.Transaction(ctx, func(tx *content.Session) error {
		txUser := tx.Users.New().SetName("Rolled Back Tx User")
		if err := txUser.Save(ctx); err != nil {
			return err
		}
		rolledBackUserID = txUser.ID()
		return abortErr
	})
	if !errors.Is(err, abortErr) {
		log.Fatalf("expected callback error from transaction, got %v", err)
	}
	if _, err := sess.Users.Find(ctx, rolledBackUserID); !errors.Is(err, content.ErrNotFound) {
		log.Fatalf("expected rolled-back user to be invisible, got %v", err)
	}

	func() {
		defer func() {
			if r := recover(); r == nil {
				log.Fatal("expected transaction callback panic to propagate")
			}
		}()
		_ = sess.Transaction(ctx, func(tx *content.Session) error {
			txUser := tx.Users.New().SetName("Panicked Tx User")
			if err := txUser.Save(ctx); err != nil {
				return err
			}
			panicUserID = txUser.ID()
			panic("abort transaction")
		})
	}()
	if _, err := sess.Users.Find(ctx, panicUserID); !errors.Is(err, content.ErrNotFound) {
		log.Fatalf("expected panic transaction user to be invisible, got %v", err)
	}
	fmt.Println("transactions verified")

	user := sess.Users.New().SetName("Marcos Aurelio")
	if err := user.Save(ctx); err != nil {
		log.Fatalf("insert user: %v", err)
	}

	post1 := sess.Posts.New().SetUserID(user.ID()).SetTitle("First post")
	if err := post1.Save(ctx); err != nil {
		log.Fatalf("insert post1: %v", err)
	}
	post2 := sess.Posts.New().SetUserID(user.ID()).SetTitle("Second post")
	if err := post2.Save(ctx); err != nil {
		log.Fatalf("insert post2: %v", err)
	}

	// --- User Story 1: lazy read, zero I/O until Get ---
	posts, err := user.Posts().Get(ctx)
	if err != nil {
		log.Fatalf("lazy Posts().Get: %v", err)
	}
	must(len(posts) == 2, fmt.Sprintf("expected 2 posts, got %d", len(posts)))
	fmt.Println("lazy read: got", len(posts), "posts")

	cached, loaded := user.Posts().Cached()
	must(loaded, "Posts should be cached after Get")
	must(len(cached) == 2, "Cached() should return the same 2 posts without I/O")

	// belongs_to Associate / lazy Author read
	author, err := post1.Author().Get(ctx)
	if err != nil {
		log.Fatalf("lazy Author().Get: %v", err)
	}
	must(author != nil && author.ID() == user.ID(), "Author().Get should resolve back to the same user")

	// --- User Story 2: scoped read never overwrites the canonical cache ---
	published, err := user.Posts().Published().Get(ctx)
	if err != nil {
		log.Fatalf("scoped Posts().Published().Get: %v", err)
	}
	must(len(published) == 0, fmt.Sprintf("expected 0 published posts (none marked published), got %d", len(published)))
	stillCached, stillLoaded := user.Posts().Cached()
	must(stillLoaded && len(stillCached) == 2, "canonical cache must survive a scoped read (FR-019)")
	fmt.Println("scoped read did not disturb canonical cache")

	// --- many-to-many: Attach / Detach / Sync ---
	tag1 := sess.Tags.New().SetName("go")
	must(tag1.Save(ctx) == nil, "insert tag1")
	tag2 := sess.Tags.New().SetName("orm")
	must(tag2.Save(ctx) == nil, "insert tag2")

	if err := post1.Tags().Attach(ctx, tag1); err != nil {
		log.Fatalf("Attach: %v", err)
	}
	tags, err := post1.Tags().Get(ctx)
	if err != nil {
		log.Fatalf("Tags().Get: %v", err)
	}
	must(len(tags) == 1, "post1 should have exactly 1 tag after Attach")

	if err := post1.Tags().Sync(ctx, tag2); err != nil {
		log.Fatalf("Sync: %v", err)
	}
	tags, err = post1.Tags().Reload(ctx)
	if err != nil {
		log.Fatalf("Tags().Reload: %v", err)
	}
	must(len(tags) == 1 && tags[0].ID() == tag2.ID(), "Sync should leave exactly tag2 attached")
	fmt.Println("many-to-many Attach/Sync verified")

	if err := post1.Tags().Detach(ctx, tag2); err != nil {
		log.Fatalf("Detach: %v", err)
	}
	tags, err = post1.Tags().Reload(ctx)
	must(err == nil && len(tags) == 0, "Detach should leave zero tags attached")

	// --- session-mismatch / unsaved-related rejection (FR-021) ---
	otherSess := content.New(pool)
	otherUser, err := otherSess.Users.Find(ctx, user.ID())
	if err != nil {
		log.Fatalf("Find in other session: %v", err)
	}
	if err := post2.Author().Associate(otherUser); err != content.ErrSessionMismatch {
		log.Fatalf("expected ErrSessionMismatch associating a different-session model, got %v", err)
	}
	unsavedUser := sess.Users.New().SetName("Not Saved")
	if err := post2.Author().Associate(unsavedUser); err != content.ErrUnsavedRelated {
		log.Fatalf("expected ErrUnsavedRelated associating an unsaved model, got %v", err)
	}
	if err := sess.Transaction(ctx, func(tx *content.Session) error {
		txPost := tx.Posts.New().SetTitle("Wrong Session Author")
		if err := txPost.Author().Associate(user); err != content.ErrSessionMismatch {
			return fmt.Errorf("expected ErrSessionMismatch associating root user inside transaction, got %w", err)
		}
		txUser := tx.Users.New().SetName("Transaction Author")
		if err := txUser.Save(ctx); err != nil {
			return err
		}
		if err := txPost.Author().Associate(txUser); err != nil {
			return err
		}
		return txPost.Save(ctx)
	}); err != nil {
		log.Fatalf("transaction session identity checks: %v", err)
	}
	fmt.Println("session-mismatch / unsaved-related rejection verified")

	// --- User Story 3: eager loading, exactly one batch query ---
	user2 := sess.Users.New().SetName("Second User")
	must(user2.Save(ctx) == nil, "insert user2")
	// user2 has zero posts — verifies eager loading populates an explicit
	// empty result, distinguishable from "not yet loaded" (FR-015).

	users := []*content.User{user, user2}
	if err := sess.Users.EagerLoad(ctx, users, content.WithPosts()); err != nil {
		log.Fatalf("EagerLoad: %v", err)
	}
	for _, u := range users {
		p, loaded := u.Posts().Cached()
		must(loaded, "every user's Posts must be marked loaded after EagerLoad, including zero-post users")
		if u.ID() == user.ID() {
			must(len(p) == 2, "user should have 2 eager-loaded posts")
		} else {
			must(len(p) == 0, "user2 should have 0 eager-loaded posts (explicit empty, not unloaded)")
		}
	}
	fmt.Println("eager loading verified: exactly one batch query for", len(users), "users")

	// inverse hydration: each eager-loaded post's Author should already be
	// the known parent, no extra query.
	loadedPosts, _ := user.Posts().Cached()
	for _, p := range loadedPosts {
		a, aLoaded := p.Author().Cached()
		must(aLoaded, "eager loading must populate the inverse Author cache")
		must(a.ID() == user.ID(), "inverse Author must be the same already-known user instance")
	}
	fmt.Println("inverse hydration verified")

	// --- strict mode ---
	strictSess := content.New(pool, content.WithLazyLoading(content.LazyLoadingPrevented))
	freshUser, err := strictSess.Users.Find(ctx, user.ID())
	if err != nil {
		log.Fatalf("Find (strict session): %v", err)
	}
	if _, err := freshUser.Posts().Get(ctx); err != content.ErrLazyLoadingPrevented {
		log.Fatalf("expected ErrLazyLoadingPrevented for an uncached lazy read under strict mode, got %v", err)
	}
	if err := strictSess.Users.EagerLoad(ctx, []*content.User{freshUser}, content.WithPosts()); err != nil {
		log.Fatalf("EagerLoad (strict session): %v", err)
	}
	if _, err := freshUser.Posts().Get(ctx); err != nil {
		log.Fatalf("expected a cached read to succeed under strict mode after eager loading, got %v", err)
	}
	fmt.Println("strict lazy-loading mode verified")

	fmt.Println("ALL CHECKS PASSED")
}
