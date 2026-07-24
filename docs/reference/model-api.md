# Model API reference

The examples use `User`, but the generator applies the same contract to each configured model.

## Getters

```go
func (u *User) ID() UserID
func (u *User) Name() string
func (u *User) Email() string
func (u *User) Active() bool
func (u *User) CreatedAt() time.Time
func (u *User) UpdatedAt() time.Time
```

Only readable fields receive getters.

## Chainable setters

```go
func (u *User) SetName(value string) *User
func (u *User) SetEmail(value string) *User
func (u *User) SetActive(value bool) *User
```

Only mutable or fillable fields receive setters. A setter compares the new current value with the original snapshot and updates dirty state deterministically.

## Original values

```go
func (u *User) OriginalName() string
func (u *User) OriginalEmail() string
func (u *User) OriginalActive() bool
```

## Dirty state

```go
func (u *User) IsDirty(fields ...UserField) bool
func (u *User) IsClean(fields ...UserField) bool
func (u *User) DirtyFields() []UserField
```

`DirtyFields` follows model declaration order rather than map iteration order.

## Validation

```go
func (u *User) Err() error
func (u *User) HasErrors() bool
func (u *User) FieldError(field UserField) error
func (u *User) ClearErrors() *User
```

Protected package-level helpers support handwritten behavior:

```go
func (u *User) setFieldError(field UserField, err error) *User
func (u *User) clearFieldError(field UserField) *User
```

## Lifecycle state

```go
func (u *User) Exists() bool
func (u *User) IsNew() bool
func (u *User) IsDeleted() bool
func (u *User) IsAttached() bool
```

## Terminal methods

```go
func (u *User) Save(ctx context.Context) error
func (u *User) Delete(ctx context.Context) error
func (u *User) Refresh(ctx context.Context) error
```

`Save` validates first, inserts new models, skips clean persisted models, updates dirty persisted models, and rejects deleted or detached models.

## Non-exported lifecycle methods

Hydration, snapshot synchronization, store attachment, and persisted-state mutation are private to the bounded-context package. Public `Restore` or `MarkPersisted` methods are not generated.
