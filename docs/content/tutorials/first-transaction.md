# Tutorial: run model operations in a transaction

## Goal

Use a transaction-bound model session so every model operation in the callback executes through sqlc queries bound to the same pgx transaction.

## 1. Start from the root session

```go
models := content.New(pool)
```

## 2. Open a transaction

```go
err := models.Transaction(ctx, func(txModels *content.Session) error {
    user, err := txModels.Users.Find(ctx, userID)
    if err != nil {
        return err
    }

    post := txModels.Posts.New().
        SetTitle("Transaction-safe post").
        SetBody(body)

    post.Author().Associate(user)

    if err := post.Save(ctx); err != nil {
        return err
    }

    return user.Activate().Save(ctx)
})
```

The callback receives a new session with:

- sqlc queries rebound through `WithTx`;
- a distinct session identity;
- fresh collections;
- the same runtime options unless overridden.

## 3. Understand commit and rollback

The transaction commits only when the callback returns `nil`.

An error causes rollback:

```go
err := models.Transaction(ctx, func(txModels *content.Session) error {
    user, err := txModels.Users.Find(ctx, userID)
    if err != nil {
        return err
    }

    return user.Suspend("manual review").Save(ctx)
})
```

Rollback is registered with `defer` immediately after the transaction begins. This also protects panic paths from leaking an open transaction.

## 4. Do not reuse outer-session models

This is invalid:

```go
user, err := models.Users.Find(ctx, userID)
if err != nil {
    return err
}

return models.Transaction(ctx, func(txModels *content.Session) error {
    return user.Save(ctx)
})
```

`user` remains attached to the root session and would persist outside the transaction. The framework does not silently reattach it.

Load or create all participating models through `txModels`.

## 5. Observe association safety

The following is rejected:

```go
outerUser, _ := models.Users.Find(ctx, userID)

return models.Transaction(ctx, func(txModels *content.Session) error {
    post := txModels.Posts.New()
    post.Author().Associate(outerUser)
    return post.Save(ctx)
})
```

The association records `ErrSessionMismatch` because the models belong to different sessions.

## 6. Unsaved relation policy

The first release rejects an unsaved related model whose identifier is unavailable:

```go
user := txModels.Users.New().SetName("Marcos")
post := txModels.Posts.New().SetTitle("Example")

post.Author().Associate(user) // ErrUnsavedRelatedModel when user has no key
```

Save the parent first, or use application-generated identifiers. Automatic cascade persistence is deliberately deferred.
