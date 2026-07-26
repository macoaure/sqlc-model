# Repository Guidelines

## Project Structure & Module Organization

`sqlc-model` is a Go module for generating rich model code on top of sqlc. The CLI entrypoint lives in `cmd/sqlc-model/`. Internal implementation packages are under `internal/`, grouped by responsibility such as `config`, `contract`, `generate`, `plan`, `relation`, and `diagnostics`.

Tests are split by level: `tests/unit/` for focused package behavior, `tests/golden/` for generator snapshots and diagnostics, `tests/compile/` for standalone compile fixtures, and `tests/integration/` for PostgreSQL-backed behavior. Golden output fixtures live in `testdata/golden/`. Documentation is in `docs/content/`, organized with Diataxis-style tutorials, how-to guides, reference, and explanation pages.

## Build, Test, and Development Commands

- `go mod download`: downloads module dependencies.
- `go build ./...`: verifies all packages compile.
- `go test ./...`: runs the normal test suite.
- `go test -race ./...`: runs tests with the race detector.
- `go test ./tests/compile/...`: checks generated fixture modules compile.
- `go test ./tests/golden`: runs generator snapshot tests.
- `SQLC_RICHMODEL_TEST_DATABASE_URL='postgres://user:pass@localhost:5432/postgres?sslmode=disable' go test ./tests/integration/...`: runs database integration tests.
- `golangci-lint run`: runs lint checks used by CI.

## Coding Style & Naming Conventions

Use standard Go formatting: run `gofmt`/`go fmt ./...` before committing. Keep package names short, lowercase, and domain-specific. Test files use Go conventions, ending in `_test.go`; generated fixture files commonly end in `_gen.go`. Prefer existing package boundaries in `internal/` over adding new layers.

## Testing Guidelines

Add the smallest test that protects the behavior changed. Use unit tests for local logic, golden tests for generator output, compile fixtures for generated API contracts, and integration tests only when PostgreSQL behavior matters. Name fixtures after the behavior boundary, and document unsupported matrix cases in `tests/compile/matrix.md`.

## Commit & Pull Request Guidelines

Recent history uses gitmoji plus Conventional Commits, for example `🐛 fix(session): preserva identidade entre sessões` and `📝 docs(readme): atualiza visão geral da biblioteca`. Keep subjects under 72 characters and use a clear scope.

Before opening a PR, run `go build ./...`, `go test ./...`, and `golangci-lint run`. PRs should target `main`, describe the change, link related issues, and note any test coverage or integration setup needed.
