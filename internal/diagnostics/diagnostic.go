package diagnostics

import "sort"

// Severity classifies a Diagnostic as blocking (Error) or advisory (Warning).
type Severity uint8

const (
	SeverityWarning Severity = iota
	SeverityError
)

func (s Severity) String() string {
	if s == SeverityError {
		return "error"
	}
	return "warning"
}

// Diagnostic reports a single problem found by any validation stage in the
// pipeline. See contracts/diagnostics-contract.md for the authoritative
// shape and failure semantics.
type Diagnostic struct {
	Severity Severity
	Path     string
	Context  string
	Model    string
	Relation string
	Query    string
	Message  string
	Hint     string
}

// Sort orders diagnostics deterministically by (Path, Context, Model, Query)
// ascending, per contracts/diagnostics-contract.md's Ordering section. It
// sorts in place and also returns the slice for convenience.
func Sort(diags []Diagnostic) []Diagnostic {
	sort.SliceStable(diags, func(i, j int) bool {
		a, b := diags[i], diags[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Context != b.Context {
			return a.Context < b.Context
		}
		if a.Model != b.Model {
			return a.Model < b.Model
		}
		if a.Relation != b.Relation {
			return a.Relation < b.Relation
		}
		return a.Query < b.Query
	})
	return diags
}

// HasError reports whether diags contains at least one SeverityError
// diagnostic. Any single error means zero files are emitted for the run
// (FR-017) — this is the atomicity check used by internal/plan and the
// top-level generation pipeline.
func HasError(diags []Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}
