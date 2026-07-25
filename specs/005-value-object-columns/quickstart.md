# Quickstart: Value Object Field Mapping

## Write the value object

```go
type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	if !strings.Contains(value, "@") {
		return Email{}, fmt.Errorf("invalid email")
	}
	return Email{value: value}, nil
}

func (e Email) String() string {
	return e.value
}
```

## Configure the field

```json
{
  "version": 1,
  "sqlc": {
    "package": "db",
    "import": "example.com/app/internal/db",
    "driver": "pgx/v5"
  },
  "contexts": [
    {
      "name": "accounts",
      "package": "accounts",
      "directory": "internal/accounts",
      "models": {
        "User": {
          "row": "User",
          "operations": {
            "find": "GetUser",
            "insert": "CreateUser",
            "update": "UpdateUser"
          },
          "fields": {
            "id": { "readable": true, "generated": "insert" },
            "email": {
              "readable": true,
              "fillable": true,
              "mutable": true,
              "value_object": {
                "type": "Email",
                "constructor": "NewEmail",
                "accessor": "String"
              }
            }
          }
        }
      }
    }
  ]
}
```

## Use the generated model

```go
email, err := NewEmail("ada@example.com")
if err != nil {
	return err
}

user := users.New(email)
if err := user.Save(ctx); err != nil {
	return err
}
```

Expected checks:
- generated `User.Email()` returns `Email`;
- generated persistence passes `Email.String()` to the sqlc-backed query;
- loading a row with an invalid email returns an error naming `User.Email`;
- a field with a custom type but no `value_object` mapping fails generation unless it is otherwise directly assignable or already recognized.
