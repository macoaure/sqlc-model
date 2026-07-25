# Implementation Plan: Value Object Field Mapping

**Branch**: `000-project-bootstrap` | **Date**: 2026-07-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-value-object-columns/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Allow a field policy to explicitly map one generated model field to a developer-owned value object type. The generator will keep the persisted/query type resolved from sqlc metadata, expose the configured value object type on the model, hydrate by calling the configured constructor, and persist by calling the configured accessor. Unconfigured type mismatches stay errors instead of falling through to nullable-wrapper guesses.

## Technical Context

**Language/Version**: Go 1.25.0 (current `go.mod`)

**Primary Dependencies**: Existing sqlc process-plugin stack (`github.com/sqlc-dev/plugin-sdk-go`), `golang.org/x/tools/imports`, stdlib, and generated-code dependency on `github.com/jackc/pgx/v5`/`pgxpool`; no new module dependency.

**Storage**: N/A for the generator; generated code targets PostgreSQL through pgx. Value-object conversion is in-memory plumbing around sqlc query parameters and result rows.

**Testing**: `go test ./...`; extend config/mapping/plan unit tests, golden output snapshots, and compile fixtures for value-object hydration, persistence conversion, constructor errors, multiple mapped fields, and unconfigured mismatches.

**Target Platform**: Cross-platform Go CLI/plugin binary invoked by `sqlc generate`; generated code targets Go server applications using pgx.

**Project Type**: Single Go module: sqlc code-generation plugin plus generated runtime surface.

**Performance Goals**: No reflection or registry. Hydration pays one constructor call per mapped field, persistence pays one accessor call per mapped query parameter.

**Constraints**: Value object source stays developer-owned. Baseline scope is non-nullable single-column value objects only. The plugin can validate mapping configuration and emitted conversion shape, but existence/signature of handwritten Go symbols is ultimately enforced by compiling generated code with the developer package.

**Scale/Scope**: Applies per configured field across all generated models. Covers FR-001 through FR-011 only; no composite value objects, nullable value-object wrappers, generated value-object source, runtime validation registry, or static analysis of arbitrary handwritten behavior.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` is still the unfilled template. There are no ratified principles or gates to evaluate, so this check has nothing enforceable and cannot fail.

## Project Structure

### Documentation (this feature)

```text
specs/005-value-object-columns/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── value-object-field-api.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 output from /speckit-tasks; not created here
```

### Source Code (repository root)

```text
cmd/
└── sqlc-gen-richmodel/       # unchanged plugin entrypoint

internal/
├── config/                   # add field.value_object decoding and validation
├── mapping/                  # keep persisted sqlc type resolution; no guessing
├── plan/                     # carry exposed type plus persisted type/converters
├── codegen/                  # render constructor/accessor conversion in record/store/model paths
├── generate/                 # unchanged file orchestration unless new diagnostics need threading
├── contract/                 # unchanged query contract validation
└── diagnostics/              # unchanged diagnostic type

tests/
├── unit/                     # config/mapping/plan/codegen assertions
├── golden/                   # generated output snapshots and regeneration checks
└── compile/                  # developer-owned value object package fixtures
```

**Structure Decision**: No new package or dependency. The smallest working change is to extend existing `FieldPolicy`, `mapping.ResolvedField`, and `plan.ResolvedField` so codegen can see both the exposed model type and the persisted sqlc type. `internal/codegen/model.go`, `record.go`, and `store.go` then render explicit conversion only for fields with `value_object`.

## Post-Design Constitution Check

Still no ratified constitution to check. Phase 1 artifacts add no external dependency and keep the implementation inside the existing generator pipeline.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations.
