# Explanation: lazy and eager loading

## Lazy loading

A relation builder performs no SQL until `Get(ctx)` or `Reload(ctx)` is called:

```go
posts, err := user.Posts().Get(ctx)
```

This provides local ergonomics while keeping I/O visible through a context and error return.

## Canonical cache

The initial implementation caches only the default unconstrained relation. A scoped result such as `Published().Limit(10)` does not overwrite it.

Caching arbitrary variants would require canonical key serialization, memory bounds, invalidation rules, and synchronization. That complexity is deferred.

## N+1 risk

Lazy access inside a parent loop can execute one query per parent. Eager loading is therefore a first-class feature, not an optimization afterthought.

```go
users, err := models.Users.Query().WithPosts().Get(ctx)
```

This requires a configured batch query. The framework does not automatically rewrite the lazy query.

## Inverse hydration

When loading posts for a known user, each post's configured `Author` cache should reference the already known user instance. This prevents an immediate redundant query and improves local coherence.

## Strict mode

A session can log or prevent uncached lazy loading. This supports development-time detection of accidental N+1 behavior without removing lazy loading from the API.
