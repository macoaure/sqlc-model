# How to handle fluent validation

## Store field errors by field

The runtime should use replacement semantics:

```go
validationErrors map[UserField]error
```

An invalid value sets or replaces the error:

```go
func (u *User) ChangeEmail(value string) *User {
    email, err := ParseEmail(value)
    if err != nil {
        return u.setFieldError(UserFieldEmail, err)
    }

    u.clearFieldError(UserFieldEmail)
    return u.SetEmail(email.String())
}
```

## Allow correction in the same chain

```go
user.
    ChangeEmail("invalid").
    ChangeEmail("valid@example.com")
```

The second call clears the first field error. `user.Err()` must be clean when no other validation errors remain.

## Validate before persistence

`Save(ctx)` checks validation before selecting insert or update behavior:

```go
if err := u.Validate(); err != nil {
    return err
}
```

No database query runs when validation fails.

## Add cross-field validation

```go
func (u *User) Validate() error {
    if u.active && u.email == "" {
        return errors.Join(u.Err(), ErrActiveUserRequiresEmail)
    }

    return u.Err()
}
```

Cross-field errors should not be attached to an arbitrary field unless the application can present them meaningfully.

## Expose inspection APIs

Generated models should provide:

```go
func (u *User) Err() error
func (u *User) HasErrors() bool
func (u *User) FieldError(field UserField) error
func (u *User) ClearErrors() *User
```

`ClearErrors` is an explicit escape hatch and should not modify current values or dirty state.
