# Contract: Value Object Field API

## Configuration

A field policy may declare a value-object mapping:

```json
{
  "readable": true,
  "fillable": true,
  "mutable": true,
  "value_object": {
    "type": "Email",
    "constructor": "NewEmail",
    "accessor": "String"
  }
}
```

Contract:
- `type` names the generated model field type;
- `constructor` is called with the persisted sqlc field value during hydration;
- `accessor` is called on the value object when a query parameter needs the persisted value;
- absent `value_object` means the existing field mapping rules apply.

## Generated Model Surface

For a `User.email` field mapped to `Email`:

```go
func (u *User) Email() Email
func (u *User) SetEmail(value Email) *User
func (u *User) OriginalEmail() Email
```

Contract:
- public model APIs expose the value object type, not the persisted primitive;
- dirty checks compare exposed field values as they do for existing fields;
- handwritten value-object validation stays outside generated code.

## Hydration

Generated row hydration must behave as if:

```go
email, err := NewEmail(row.Email)
if err != nil {
    return nil, fmt.Errorf("User.Email: %w", err)
}
model.current.Email = email
model.original.Email = email
```

Contract:
- constructor failure returns an error before the model is usable;
- the surfaced error identifies model and field;
- the original constructor error remains wrapped.

## Persistence

Generated query parameter conversion must behave as if:

```go
arg := rec.Email.String()
```

Contract:
- accessor conversion is emitted only for explicit value-object mappings;
- direct and recognized built-in mappings keep current parameter behavior;
- unconfigured mismatches fail generation instead of guessing nullable-wrapper conversion.
