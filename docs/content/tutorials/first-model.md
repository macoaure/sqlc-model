# Tutorial: generate the first rich model

## Goal

Create a `User` model backed by PostgreSQL and sqlc. At the end, the application can write:

```go
user := models.Users.New().
    SetName("Marcos Aurelio").
    SetEmail("marcos@example.com").
    Activate()

if err := user.Save(ctx); err != nil {
    return err
}
```

## 1. Define the schema

Create `database/schema/users.sql`:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

The database owns the identifier and timestamps. The public model constructor must not require these values.

## 2. Define sqlc queries

Create `database/queries/users.sql`:

```sql
-- name: GetUser :one
SELECT id, name, email, active, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email, active)
VALUES (sqlc.arg(name), sqlc.arg(email), sqlc.arg(active))
RETURNING id, name, email, active, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET
    name = sqlc.arg(name),
    email = sqlc.arg(email),
    active = sqlc.arg(active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id, name, email, active, created_at, updated_at;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;
```

Insert and update return the canonical persisted row. This is mandatory for the default lifecycle policy because database defaults, generated identifiers, timestamps, triggers, and normalization must be applied back to the same model instance.

## 3. Configure sqlc and richmodel

Create `sqlc.yaml`:

```yaml
version: "2"

plugins:
  - name: richmodel
    process:
      cmd: sqlc-model

sql:
  - name: application
    engine: postgresql
    schema:
      - database/schema
    queries:
      - database/queries

    gen:
      go:
        package: sqlcdb
        out: internal/database/sqlc
        sql_package: pgx/v5
        emit_interface: true
        query_parameter_limit: 0

    codegen:
      - plugin: richmodel
        out: internal/models
        options:
          version: 1
          sqlc:
            package: sqlcdb
            import: example.com/project/internal/database/sqlc
            driver: pgx/v5
          contexts:
            - name: content
              package: content
              directory: content
              models:
                User:
                  row: User
                  operations:
                    find: GetUser
                    insert: CreateUser
                    update: UpdateUser
                    delete: DeleteUser
                  fields:
                    id:
                      generated: insert
                      mutable: false
                    name:
                      fillable: true
                    email:
                      fillable: true
                    active:
                      fillable: true
                    created_at:
                      generated: insert
                      mutable: false
                    updated_at:
                      generated: save
                      mutable: false
```

`query_parameter_limit: 0` keeps sqlc method parameters structurally consistent by generating parameter structs rather than changing signatures based on argument count.

## 4. Generate code

Run:

```bash
sqlc generate
```

Expected output:

```text
internal/
├── database/sqlc/
│   ├── db.go
│   ├── models.go
│   ├── querier.go
│   └── users.sql.go
└── models/content/
    ├── session_gen.go
    ├── user_gen.go
    ├── user_collection_gen.go
    ├── user.go
    └── internal/user/
        ├── store_gen.go
        └── record_gen.go
```

The generator may create `user.go` only when it does not already exist. It never overwrites the file.

## 5. Create a model session

```go
models := content.New(pool)
```

The session owns sqlc queries, a unique session identity, model collections, transaction creation, lazy-loading policy, and database-error translation.

## 6. Create and save a user

```go
user := models.Users.New().
    SetName("Marcos Aurelio").
    SetEmail("marcos@example.com").
    SetActive(true)

if err := user.Save(ctx); err != nil {
    return err
}
```

After a successful insert:

```go
user.Exists()    // true
user.IsNew()     // false
user.IsDirty()   // false
user.ID()        // generated UUID
user.CreatedAt() // database value
user.UpdatedAt() // database value
```

The adapter applies the `CreateUser` return value to the existing object and synchronizes the original snapshot.

## 7. Retrieve and update

```go
user, err := models.Users.Find(ctx, userID)
if err != nil {
    return err
}

user.
    SetName("Marcos A.").
    SetEmail("marcos.a@example.com")

if err := user.Save(ctx); err != nil {
    return err
}
```

A successful update leaves the model clean and synchronizes original values.

## 8. Add handwritten behavior

Edit `internal/models/content/user.go`:

```go
package content

import "strings"

func (u *User) Rename(name string) *User {
    name = strings.TrimSpace(name)
    if name == "" {
        return u.setFieldError(UserFieldName, ErrUserNameRequired)
    }

    u.clearFieldError(UserFieldName)
    return u.SetName(name)
}

func (u *User) Activate() *User {
    return u.SetActive(true)
}

func (u *User) Deactivate() *User {
    return u.SetActive(false)
}
```

Application code now expresses intent:

```go
return user.
    Rename("Marcos Aurelio").
    Activate().
    Save(ctx)
```

The model layer is now active while sqlc remains the only persistence engine.
