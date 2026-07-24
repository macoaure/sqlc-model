# How to use value objects

## Objective

Expose a richer model type while preserving the sqlc-generated persistence type.

## Define the value object

```go
package content

type Email struct {
    value string
}

func ParseEmail(value string) (Email, error) {
    // Normalize and validate.
}

func (e Email) String() string {
    return e.value
}
```

## Configure conversion

```yaml
fields:
  email:
    value_object:
      type: Email
      from_sqlc: ParseEmail
      to_sqlc: String
```

## Generated hydration

```go
email, err := ParseEmail(row.Email)
if err != nil {
    return nil, fmt.Errorf("hydrate User.email: %w", err)
}
```

## Generated persistence conversion

```go
params.Email = user.email.String()
```

## Avoid implicit wrapper assumptions

The generator must not treat every type mismatch as a pgtype wrapper with `.Valid` and a value field. Conversion is valid only when types are directly assignable, covered by a known supported mapping, or explicitly configured.

## Keep value-object behavior handwritten

The generator emits conversion calls and field plumbing. Validation rules and domain behavior remain in developer-owned source files.
