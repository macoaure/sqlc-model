# Explanation: static query composition

An Eloquent query builder can compose arbitrary predicates and joins at runtime. sqlc starts from statically declared SQL and generates a fixed Go method contract.

Allowing this:

```go
posts.Where("score", ">", 10).OrderBy("created_at", "desc")
```

would require a second SQL engine responsible for identifier quoting, operator validation, dialect behavior, joins, parameter binding, and result inference. It would also bypass part of sqlc's static-analysis value.

The project therefore supports only typed scopes backed by known sqlc capabilities:

```go
user.Posts().Published().Latest().Limit(10).Get(ctx)
```

Each method maps to one of:

- a known query parameter and constant value;
- a known query parameter and typed argument;
- a configured query variant;
- an eager-load plan.

The generator validates combinations before source emission. If a chain cannot resolve to a statically configured query contract, it is not generated.

This constraint is the central differentiator: model ergonomics improve without obscuring which declared SQL method will execute.
