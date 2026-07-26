package content

import "testing"

func TestIntegrationFixtureExposesRelationAndTransactionAPI(t *testing.T) {
	sess := New(nil)
	if sess.Users == nil || sess.Posts == nil || sess.Tags == nil {
		t.Fatal("expected generated collections on integration session")
	}

	_ = (*Session).Transaction
	_ = WithPosts
	_ = sess.Users.New().Posts
	_ = sess.Posts.New().Author
	_ = sess.Posts.New().Tags
}
