# Relation API reference

## Has-many

```go
func (u *User) Posts() UserPostsRelation
```

```go
func (r UserPostsRelation) Get(ctx context.Context) ([]*Post, error)
func (r UserPostsRelation) Reload(ctx context.Context) ([]*Post, error)
func (r UserPostsRelation) Cached() ([]*Post, bool)
func (r UserPostsRelation) Loaded() bool
func (r UserPostsRelation) Forget() *User
func (r UserPostsRelation) New() *Post
```

Configured scopes return relation values:

```go
func (r UserPostsRelation) Published() UserPostsRelation
func (r UserPostsRelation) Latest() UserPostsRelation
func (r UserPostsRelation) Limit(int32) UserPostsRelation
```

## Belongs-to

```go
func (p *Post) Author() PostAuthorRelation
```

```go
func (r PostAuthorRelation) Get(ctx context.Context) (*User, error)
func (r PostAuthorRelation) Reload(ctx context.Context) (*User, error)
func (r PostAuthorRelation) Cached() (*User, bool)
func (r PostAuthorRelation) Loaded() bool
func (r PostAuthorRelation) Forget() *Post
func (r PostAuthorRelation) Associate(*User) *Post
func (r PostAuthorRelation) Dissociate() *Post
```

`Dissociate` exists only for nullable relations.

## Has-one

Has-one shares single-value cache behavior with belongs-to but derives the target from the parent key.

## Many-to-many

```go
func (r PostTagsRelation) Get(ctx context.Context) ([]*Tag, error)
func (r PostTagsRelation) Attach(ctx context.Context, tags ...*Tag) error
func (r PostTagsRelation) Detach(ctx context.Context, tags ...*Tag) error
func (r PostTagsRelation) Sync(ctx context.Context, tags ...*Tag) error
```

Mutation methods are generated only when explicit sqlc queries are configured.

## Cache semantics

Only the canonical unconstrained relation is cached by default. Scoped results do not overwrite canonical caches. `Cached()` never performs I/O. `Reload()` clears and reloads. `Forget()` clears without I/O.

## Builder semantics

Relation builders use value receivers. Scope calls do not mutate the parent model or another builder instance.
