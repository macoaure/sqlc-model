# Collection API reference

A collection is the Go equivalent of the static model entry point found in class-oriented Active Record APIs.

## Create a new model

```go
func (c *UserCollection) New() *User
```

The returned model is attached to the collection session and does not yet exist in the database.

## Find by identifier

```go
func (c *UserCollection) Find(
    ctx context.Context,
    id UserID,
) (*User, error)
```

No row returns `ErrNotFound`. The collection does not expose a panic-based `FindOrFail` in the baseline API.

## Named retrieval methods

Configured compatible sqlc queries may generate methods such as:

```go
func (c *UserCollection) FindByEmail(
    ctx context.Context,
    email string,
) (*User, error)
```

The generator never derives these methods from column names alone.

## Query entry point

```go
func (c *UserCollection) Query() UserQuery
```

`UserQuery` is a typed configuration object for predefined sqlc capabilities, not a general runtime SQL builder.

```go
users, err := models.Users.
    Query().
    Active().
    WithPosts().
    Limit(50).
    Get(ctx)
```

Every scope and eager-load option must resolve to configured query parameters or variants.
