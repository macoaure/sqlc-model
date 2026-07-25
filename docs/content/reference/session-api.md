# Session API reference

## Construction

```go
func New(pool *pgxpool.Pool, options ...SessionOption) *Session
```

A session contains model collections, database transaction capability, runtime policies, and a private identity token.

```go
type Session struct {
    Users *UserCollection
    Posts *PostCollection
    Tags  *TagCollection
}
```

## Transaction

```go
func (s *Session) Transaction(
    ctx context.Context,
    fn func(*Session) error,
) error
```

The callback session is bound to the underlying `pgx.Tx` and has a distinct session identity. Returning nil commits. Returning an error rolls back and returns that error. Panicking rolls back before the panic continues.

## Options

```go
func WithLazyLoading(mode LazyLoadingMode) SessionOption
```

## Lazy-loading modes

```go
type LazyLoadingMode uint8

const (
    LazyLoadingAllowed LazyLoadingMode = iota
    LazyLoadingLogged
    LazyLoadingPrevented
)
```

The mode applies only when a relation query would be executed. Reading a populated cache remains permitted.

## Identity contract

Models created or hydrated by a session retain that session. Association operations require matching identities. The framework does not silently reattach models.

## Concurrency

Session and collection concurrency depends on the underlying database implementation and immutable runtime options. Model instances themselves are mutable and not concurrency-safe.
