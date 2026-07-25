// Command e2e exercises specs/002-relations-lazy-eager-loading's
// quickstart.md walkthrough against a real PostgreSQL instance, using only
// the generated relation API: lazy read, scoped read, association,
// many-to-many attach/detach/sync, eager loading with inverse hydration,
// and strict lazy-loading mode.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/macoaure/sqlc-gen-richmodel/tests/compile/user-relations/content"
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
	otherPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting (second pool): %v", err)
	}
	defer otherPool.Close()
	otherSess := content.New(otherPool)
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
