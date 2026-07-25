# How to use transactions safely

## Use the transaction session

```go
err := models.Transaction(ctx, func(txModels *content.Session) error {
    user, err := txModels.Users.Find(ctx, userID)
    if err != nil {
        return err
    }

    return user.Activate().Save(ctx)
})
```

## Create all participating models inside the callback

A model retains the session that created or loaded it. An outer-session model cannot be silently adopted by the transaction.

## Associate only compatible models

```go
post := txModels.Posts.New()
user, err := txModels.Users.Find(ctx, userID)
if err != nil {
    return err
}

post.Author().Associate(user)
```

A cross-session association records `ErrSessionMismatch`.

## Rely on callback outcome

- `nil` commits.
- An error rolls back and is returned.
- A panic triggers deferred rollback cleanup and is rethrown.

## Do not perform hidden cascade saves

Save related new models explicitly in dependency order. Automatic cascades introduce hidden query ordering, cycle detection, and implicit transaction requirements and are outside the initial release.
