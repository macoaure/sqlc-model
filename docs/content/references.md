# External references

The architecture is designed against the public behavior documented by the following sources.

## Diátaxis

- Diátaxis framework: https://diataxis.fr/
- Tutorials: https://diataxis.fr/tutorials/
- How-to guides: https://diataxis.fr/how-to-guides/
- Reference: https://diataxis.fr/reference/
- Explanation: https://diataxis.fr/explanation/

## sqlc

- Documentation: https://docs.sqlc.dev/
- Plugin development: https://docs.sqlc.dev/en/latest/guides/plugins.html
- Query annotations: https://docs.sqlc.dev/en/latest/reference/query-annotations.html
- Transactions: https://docs.sqlc.dev/en/latest/howto/transactions.html
- Named parameters: https://docs.sqlc.dev/en/latest/howto/named_parameters.html
- Go generation configuration: https://docs.sqlc.dev/en/latest/reference/config.html

## Go

- Package names: https://go.dev/blog/package-names
- Context package: https://pkg.go.dev/context
- Go modules reference: https://go.dev/ref/mod

## Laravel Eloquent concepts

- Eloquent ORM: https://laravel.com/docs/eloquent
- Eloquent relationships: https://laravel.com/docs/eloquent-relationships

Laravel is a source of developer-experience concepts, not a behavioral compatibility target. Go semantics, sqlc contracts, and explicit I/O boundaries take precedence.
