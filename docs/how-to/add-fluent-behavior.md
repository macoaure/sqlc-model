# How to add fluent model behavior

## Objective

Add domain-oriented, chainable behavior without editing generated files.

## Add the method to the handwritten model file

```go
package content

import "strings"

func (u *User) Suspend(reason string) *User {
    reason = strings.TrimSpace(reason)
    if reason == "" {
        return u.setFieldError(
            UserFieldSuspensionReason,
            ErrSuspensionReasonRequired,
        )
    }

    u.clearFieldError(UserFieldSuspensionReason)

    return u.
        SetActive(false).
        SetSuspensionReason(reason)
}
```

## Preserve the fluent contract

Use a pointer receiver and return the model pointer:

```go
func (u *User) Rename(name string) *User
```

This supports:

```go
err := user.
    Rename("Marcos Aurelio").
    Activate().
    Save(ctx)
```

## Keep I/O at terminal methods

A fluent method may normalize input, validate values, update fields, mark dirty state, clear field errors, invalidate local relation caches, and compose other nonterminal methods.

It must not call `Save`, open a transaction, execute an sqlc query, trigger a lazy relation load, or switch sessions.

## Prefer intent methods over raw setters in application code

Generated setters provide mechanical mutation:

```go
user.SetActive(false)
```

Handwritten methods communicate business intent:

```go
user.Suspend("manual review")
```

Both are valid. The domain-oriented method should normally compose generated setters so dirty tracking remains centralized.

## Preserve correction semantics

When a method validates a field, clear the previous field error after a valid value is accepted. An append-only error list would leave the model permanently invalid after one failed attempt.
