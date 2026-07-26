# Test Fixtures

Generated-model tests use stable fixture names in failure output:

- `tests/golden/<name>`: generator request snapshots and diagnostics.
- `tests/compile/<name>`: standalone Go modules that must compile.
- `tests/integration/<name>`: PostgreSQL behavior checks gated by `SQLC_RICHMODEL_TEST_DATABASE_URL`.

Prefer one named fixture per behavior boundary. If a matrix combination is intentionally unsupported, document it in `tests/compile/matrix.md` instead of leaving it implicit.
