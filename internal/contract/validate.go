package contract

import (
	"fmt"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// ValidateOperation checks that queryName exists among queries and its
// command satisfies the requirement for kind (diagnostics-contract.md
// stages 4 and 6; FR-003, FR-004). It returns the matched query — nil on
// any failure — and the diagnostics produced.
func ValidateOperation(kind OperationKind, queryName string, queries []*pb.Query, path, context, model string) (*pb.Query, []diagnostics.Diagnostic) {
	req, ok := requirements[kind]
	if !ok {
		return nil, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("unknown operation kind %q", kind),
		}}
	}

	var match *pb.Query
	for _, q := range queries {
		if q.Name == queryName {
			match = q
			break
		}
	}
	if match == nil {
		return nil, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     path,
			Context:  context,
			Model:    model,
			Query:    queryName,
			Message:  fmt.Sprintf("%s operation references query %q, but no such query was found", kind, queryName),
			Hint:     "check the query name matches a `-- name:` annotation in your sqlc queries",
		}}
	}

	if match.Cmd == req.cmd {
		return match, nil
	}
	if req.fallbackCmd != "" && match.Cmd == req.fallbackCmd {
		return match, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityWarning,
			Path:     path,
			Context:  context,
			Model:    model,
			Query:    queryName,
			Message:  fmt.Sprintf("%s query %q uses %s; affected-row-count detection is unavailable without %s", kind, queryName, req.fallbackCmd, req.cmd),
			Hint:     fmt.Sprintf("change the query annotation to %s if you need affected-row-count detection", req.cmd),
		}}
	}

	hint := ""
	if kind == Insert || kind == Update {
		hint = "add RETURNING ... and change the annotation to :one, or select an explicit non-hydrating insert policy when that policy becomes supported"
	}
	return nil, []diagnostics.Diagnostic{{
		Severity: diagnostics.SeverityError,
		Path:     path,
		Context:  context,
		Model:    model,
		Query:    queryName,
		Message:  fmt.Sprintf("query %s uses %s, but the %s lifecycle requires %s so the persisted row can hydrate generated fields", queryName, match.Cmd, kind, req.cmd),
		Hint:     hint,
	}}
}
