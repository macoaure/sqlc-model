# Tutorial: add the first lazy relation

## Goal

Add `User has many Posts` with this API:

```go
posts, err := user.
    Posts().
    Published().
    Latest().
    Limit(10).
    Get(ctx)
```

## 1. Define the posts table

```sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 2. Define the lazy query

```sql
-- name: ListPostsByUser :many
SELECT id, user_id, title, body, published, created_at, updated_at
FROM posts
WHERE user_id = sqlc.arg(user_id)
  AND (
      sqlc.narg(published)::BOOLEAN IS NULL
      OR published = sqlc.narg(published)
  )
ORDER BY
    CASE WHEN sqlc.arg(order_code)::TEXT = 'created_at_asc'
         THEN created_at END ASC,
    CASE WHEN sqlc.arg(order_code)::TEXT = 'created_at_desc'
         THEN created_at END DESC
LIMIT sqlc.arg(result_limit)
OFFSET sqlc.arg(result_offset);
```

Every supported scope is represented by a declared query parameter. The model generator does not append SQL fragments dynamically.

## 3. Define the eager query

```sql
-- name: ListPostsByUserIDs :many
SELECT id, user_id, title, body, published, created_at, updated_at
FROM posts
WHERE user_id = ANY(sqlc.arg(user_ids)::UUID[])
ORDER BY user_id, created_at DESC;
```

The batch query is separate because the generator does not transform a single-parent query into a multi-parent query.

## 4. Configure the relation

```yaml
models:
  User:
    relations:
      Posts:
        kind: has_many
        model: Post
        local_key: id
        foreign_key: user_id
        inverse: Author
        lazy_query: ListPostsByUser
        eager_query: ListPostsByUserIDs
        parameters:
          user_id:
            source: parent.id
          published:
            source: scope.Published
          order_code:
            source: scope.Order
            default: created_at_desc
          result_limit:
            source: scope.Limit
            default: 100
          result_offset:
            source: scope.Offset
            default: 0
        scopes:
          Published:
            parameter: published
            value: true
          Unpublished:
            parameter: published
            value: false
          Latest:
            parameter: order_code
            value: created_at_desc
          Oldest:
            parameter: order_code
            value: created_at_asc
          Limit:
            parameter: result_limit
            argument: int32
          Offset:
            parameter: result_offset
            argument: int32
```

The generator validates query names, command kinds, parameter names, parameter types, and result fields before rendering source files.

## 5. Generate and load

```bash
sqlc generate
```

Public output remains in one bounded-context package:

```text
internal/models/content/
├── user_gen.go
├── user_posts_relation_gen.go
├── post_gen.go
├── post_author_relation_gen.go
└── internal/
    ├── user/posts/
    │   ├── lazy_loader_gen.go
    │   └── eager_loader_gen.go
    └── post/author/
        ├── lazy_loader_gen.go
        └── eager_loader_gen.go
```

Use the relation:

```go
posts, err := user.Posts().Get(ctx)
```

`Posts()` creates a relation builder and performs no SQL. `Get(ctx)` is the terminal operation.

## 6. Observe canonical caching

```go
posts, err := user.Posts().Get(ctx) // database
if err != nil {
    return err
}

samePosts, err := user.Posts().Get(ctx) // canonical cache
```

A constrained relation does not replace the canonical cache:

```go
published, err := user.Posts().Published().Limit(10).Get(ctx)
```

The initial implementation does not cache constrained variants because doing so requires stable cache keys, unbounded cache management, and more complex invalidation.

## 7. Read the cache without I/O

```go
posts, loaded := user.Posts().Cached()
```

`Cached()` never performs SQL. A loaded empty slice is distinguishable from an unloaded relation.

## 8. Reload or forget

```go
posts, err := user.Posts().Reload(ctx)
```

`Reload` clears the canonical cache and performs a fresh query.

```go
user.Posts().Forget()
```

`Forget` clears the cache without performing I/O and returns the parent model for chaining.
