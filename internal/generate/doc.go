// Package generate is the thin orchestrator chaining internal/config ->
// internal/plan -> internal/codegen into the plugin's full request/response
// cycle, and the FR-017 atomicity check that decides whether any output is
// emitted at all for a given run.
package generate
