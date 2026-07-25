package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"

	"golang.org/x/tools/imports"
)

// render executes a text/template, then formats and import-fixes the
// result (research.md "Code generation rendering strategy"). Any failure —
// a template bug or output that doesn't parse as valid Go — is reported as
// a diagnostic against filePath (diagnostics-contract.md stage 8, "Go
// formatting validation") rather than as a Go error, since it's still a
// generation-time failure that must abort the run per FR-017.
func render(filePath, templateName, tmplText string, data any, context, model string) ([]byte, []diagnostics.Diagnostic) {
	tmpl, err := template.New(templateName).Funcs(templateFuncs).Parse(tmplText)
	if err != nil {
		return nil, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     filePath,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("internal error: template %s failed to parse: %v", templateName, err),
		}}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     filePath,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("internal error: rendering %s failed: %v", filePath, err),
		}}
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     filePath,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("generated file %s does not format cleanly: %v", filePath, err),
			Hint:     "this is a generator bug, not a configuration problem — please report it",
		}}
	}

	fixed, err := imports.Process(filePath, formatted, nil)
	if err != nil {
		return nil, []diagnostics.Diagnostic{{
			Severity: diagnostics.SeverityError,
			Path:     filePath,
			Context:  context,
			Model:    model,
			Message:  fmt.Sprintf("generated file %s failed import resolution: %v", filePath, err),
			Hint:     "this is a generator bug, not a configuration problem — please report it",
		}}
	}

	return fixed, nil
}

var templateFuncs = template.FuncMap{}
