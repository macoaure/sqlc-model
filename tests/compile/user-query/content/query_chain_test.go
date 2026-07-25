package content

import (
	"context"
	"testing"
)

func TestQueryChainAPISignatures(t *testing.T) {
	s := New(nil)
	q := s.Users.Query().
		Active().
		OrderByName().
		Limit(50).
		WithPosts()

	var _ UserQuery = q
	var _ func(UserQuery, context.Context) ([]*User, error) = UserQuery.Get
}
