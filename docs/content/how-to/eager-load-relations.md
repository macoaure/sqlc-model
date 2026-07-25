# How to eager-load relations

## Configure a batch query

A lazy query accepts one parent identifier. Eager loading requires a separate query that accepts many parent identifiers.

```sql
-- name: ListPostsByUserIDs :many
SELECT id, user_id, title, body, published, created_at, updated_at
FROM posts
WHERE user_id = ANY(sqlc.arg(user_ids)::UUID[])
ORDER BY user_id, created_at DESC;
```

## Configure the relation

```yaml
relations:
  Posts:
    lazy_query: ListPostsByUser
    eager_query: ListPostsByUserIDs
    eager_group_field: user_id
```

## Use eager loading

```go
users, err := models.Users.
    Query().
    WithPosts().
    Get(ctx)
```

The execution plan is:

1. execute the configured users query;
2. collect user identifiers;
3. execute `ListPostsByUserIDs` once;
4. group rows by parent key;
5. hydrate posts;
6. populate every user's canonical `Posts` cache, including loaded empty slices;
7. populate configured inverse `Author` caches with the existing user instance.

## Inspect loaded data without SQL

```go
for _, user := range users {
    posts, loaded := user.Posts().Cached()
    if !loaded {
        return errors.New("posts were not eager loaded")
    }
    _ = posts
}
```

## Configure nested eager loading

```go
users, err := models.Users.
    Query().
    WithPosts(func(posts UserPostsEagerLoad) UserPostsEagerLoad {
        return posts.WithTags()
    }).
    Get(ctx)
```

Nested eager loading is represented by an execution plan, not by public packages such as `models/user/posts/tags`.

## Prevent accidental lazy loading

In development:

```go
models := content.New(
    pool,
    content.WithLazyLoading(content.LazyLoadingPrevented),
)
```

An uncached `Get(ctx)` then returns `ErrLazyLoadingPrevented`. Eager-loaded or already cached relations remain readable.
