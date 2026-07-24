# Generated files reference

## Recommended output

```text
internal/
├── database/sqlc/
│   ├── db.go
│   ├── models.go
│   ├── querier.go
│   └── *.sql.go
└── models/
    └── content/
        ├── session_gen.go
        ├── errors_gen.go
        ├── fields_gen.go
        ├── user_gen.go
        ├── user_collection_gen.go
        ├── user_posts_relation_gen.go
        ├── user.go
        ├── post_gen.go
        ├── post_collection_gen.go
        ├── post_author_relation_gen.go
        ├── post_tags_relation_gen.go
        ├── post.go
        └── internal/
            ├── user/
            │   ├── store_gen.go
            │   ├── record_gen.go
            │   └── posts/
            │       ├── lazy_loader_gen.go
            │       └── eager_loader_gen.go
            └── post/
                ├── store_gen.go
                ├── record_gen.go
                ├── author/
                │   ├── lazy_loader_gen.go
                │   └── eager_loader_gen.go
                └── tags/
                    ├── lazy_loader_gen.go
                    └── eager_loader_gen.go
```

## Ownership rules

| Pattern | Owner | Behavior |
|---|---|---|
| `*_gen.go` | Generator | Replaced atomically on generation |
| `<model>.go` | Developer | Never overwritten |
| sqlc output | sqlc-gen-go | Never modified by richmodel |
| non-generated files | Developer | Preserved |

## Public package rule

Concrete related model types and their public relation builders live in one bounded-context package. Nested directories are reserved for internal adapters and loaders that do not import the public package.

## Atomic generation

The plugin must build, validate, format, and sort all output before returning files. A partial configuration failure must not emit a partially updated model layer.

## Deterministic order

Contexts, models, fields, operations, relations, scopes, imports, declarations, and files are sorted. Configuration maps are never iterated directly when output order matters.
