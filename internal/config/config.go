package config

import (
	"encoding/json"
	"fmt"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
)

// SqlcTarget identifies the sqlc-gen-go output the generated code will
// import (data-model.md "SqlcTarget").
type SqlcTarget struct {
	Package string
	Import  string
	Driver  string
}

// supportedDrivers is the initial supported baseline (research.md,
// contracts/config.schema.json).
var supportedDrivers = map[string]bool{
	"pgx/v5": true,
}

// RootConfiguration is the top-level `options` value passed to the plugin
// (data-model.md "RootConfiguration").
type RootConfiguration struct {
	Version  int
	Sqlc     SqlcTarget
	Contexts []BoundedContext
}

type sqlcTargetJSON struct {
	Package string `json:"package"`
	Import  string `json:"import"`
	Driver  string `json:"driver"`
}

type rootJSON struct {
	Version  int               `json:"version"`
	Sqlc     sqlcTargetJSON    `json:"sqlc"`
	Contexts []json.RawMessage `json:"contexts"`
}

// Decode parses and validates the plugin's raw `options` bytes into a
// RootConfiguration, collecting diagnostics across every structural
// validation stage. The caller (internal/generate) is responsible for
// checking diagnostics.HasError before proceeding to the next pipeline
// stage or emitting any output (FR-017).
func Decode(pluginOptions []byte) (*RootConfiguration, []diagnostics.Diagnostic) {
	var rj rootJSON
	if err := DecodeStrict(pluginOptions, &rj); err != nil {
		return nil, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     "options",
			Message:  fmt.Sprintf("invalid configuration: %v", err),
		}}
	}

	var diags []diagnostics.Diagnostic

	if rj.Version != SupportedVersion {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     "options.version",
			Message:  fmt.Sprintf("unsupported configuration schema version %d; this generator supports version %d only", rj.Version, SupportedVersion),
			Hint:     fmt.Sprintf("set version: %d", SupportedVersion),
		})
		// The version gate short-circuits before any other validation:
		// an unsupported schema shouldn't be interpreted at all (FR-008).
		return nil, diags
	}

	if rj.Sqlc.Package == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     "options.sqlc.package",
			Message:  "\"sqlc.package\" is required",
		})
	}
	if rj.Sqlc.Import == "" {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     "options.sqlc.import",
			Message:  "\"sqlc.import\" is required",
		})
	}
	if !supportedDrivers[rj.Sqlc.Driver] {
		diags = append(diags, diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Path:     "options.sqlc.driver",
			Message:  fmt.Sprintf("unsupported driver %q; supported drivers: pgx/v5", rj.Sqlc.Driver),
		})
	}

	root := &RootConfiguration{
		Version: rj.Version,
		Sqlc: SqlcTarget{
			Package: rj.Sqlc.Package,
			Import:  rj.Sqlc.Import,
			Driver:  rj.Sqlc.Driver,
		},
	}

	seenContexts := make(map[string]bool, len(rj.Contexts))
	for i, raw := range rj.Contexts {
		bc, cdiags := decodeContext(raw, i)
		diags = append(diags, cdiags...)
		if bc.Name != "" {
			if seenContexts[bc.Name] {
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Path:     fmt.Sprintf("contexts.%s", bc.Name),
					Context:  bc.Name,
					Message:  fmt.Sprintf("duplicate context name %q", bc.Name),
				})
			}
			seenContexts[bc.Name] = true
		}
		root.Contexts = append(root.Contexts, bc)
	}

	return root, diags
}
