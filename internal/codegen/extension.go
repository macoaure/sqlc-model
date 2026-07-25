package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/macoaure/sqlc-gen-richmodel/internal/plan"
)

const extensionTemplate = `package %s

// %s is created once by sqlc-gen-richmodel and never overwritten. Add
// handwritten domain methods for %s here — they will survive every future
// regeneration.
`

// ExtensionFile is a developer-owned file the generator creates exactly
// once and never overwrites (FR-015, FR-016; data-model.md "Developer-Owned
// Extension File").
type ExtensionFile struct {
	// Path is relative to the plugin's output root, e.g. "content/user.go".
	Path string
	// Contents is the minimal stub to emit — only meaningful when the file
	// does not already exist on disk.
	Contents []byte
}

// PlanExtensionFile determines whether m's extension file already exists in
// outputRoot. If it does, the returned ExtensionFile's Contents is nil and
// the caller MUST NOT include this path in the response at all — including
// it, even with identical stub contents, would overwrite whatever the
// developer has already added (FR-015). outputRoot is the plugin's
// configured `out` directory; if empty (e.g. under test, where no real
// filesystem check is meaningful) the file is always treated as absent.
func PlanExtensionFile(outputRoot string, ctx plan.ResolvedContext, m plan.ResolvedModel) (ExtensionFile, bool) {
	relPath := fmt.Sprintf("%s/%s.go", ctx.Directory, fileStem(m.Row))
	ef := ExtensionFile{Path: relPath}

	if outputRoot != "" {
		absPath := filepath.Join(outputRoot, filepath.FromSlash(relPath))
		if _, err := os.Stat(absPath); err == nil {
			// Already present on disk: emit nothing for this path.
			return ef, false
		}
	}

	ef.Contents = fmt.Appendf(nil, extensionTemplate, ctx.Package, m.Row, m.Row)
	return ef, true
}
