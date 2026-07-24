# Explanation: why Active Record over sqlc

sqlc generates typed functions around explicit SQL, but applications still need to coordinate row conversion, model validation, update parameters, generated values, relationships, transaction-bound query objects, and error translation.

Direct sqlc usage often looks like:

```go
row, err := queries.GetUser(ctx, id)
if err != nil {
    return err
}

updated, err := queries.UpdateUser(ctx, sqlcdb.UpdateUserParams{
    ID:     row.ID,
    Name:   newName,
    Email:  row.Email,
    Active: row.Active,
})
```

The rich-model layer consolidates this lifecycle:

```go
user, err := models.Users.Find(ctx, id)
if err != nil {
    return err
}

return user.Rename(newName).Save(ctx)
```

The API deliberately allows `Save`, `Delete`, and `Refresh` on the model. That is an Active Record characteristic. A public repository interface would contradict the intended developer experience and duplicate abstractions already provided by sqlc.

The implementation still contains stores and loaders, but they are private adapters. The public domain model is not persistence-ignorant; it is session-attached by design.

The project is inspired by Eloquent's ergonomics, not obligated to reproduce PHP behavior. Go's explicit errors, contexts, package model, and static query contracts define the actual architecture.
