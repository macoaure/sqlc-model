package generate

import (
	"os"
	"path/filepath"

	"github.com/macoaure/sqlc-gen-richmodel/internal/codegen"
	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/plan"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// Generate runs the full pipeline for one plugin invocation: decode +
// validate configuration, build the generation plan against sqlc's request
// metadata, then render every file. Diagnostics are collected across every
// stage and sorted before being returned; per FR-017, resp is nil whenever
// diags contains any error-severity diagnostic — the caller must not emit
// any output for the run in that case.
func Generate(req *pb.GenerateRequest) (resp *pb.GenerateResponse, diags []diagnostics.Diagnostic) {
	root, cdiags := config.Decode(req.GetPluginOptions())
	diags = append(diags, cdiags...)
	if root == nil {
		return nil, diagnostics.Sort(diags)
	}

	p, pdiags := plan.Build(root, req, diags)
	diags = pdiags // plan.Build already carries prior diagnostics forward
	if p == nil {
		return nil, diagnostics.Sort(diags)
	}

	files, rdiags := codegen.Render(p, outputRoot(req))
	diags = append(diags, rdiags...)
	if files == nil {
		return nil, diagnostics.Sort(diags)
	}

	return &pb.GenerateResponse{Files: files}, diagnostics.Sort(diags)
}

// outputRoot resolves the plugin's configured `out` directory to an
// absolute path, so internal/codegen can check whether a model's
// developer-owned extension file already exists on disk (FR-015). sqlc
// invokes process plugins with the working directory set to the project
// root containing sqlc.yaml, so `out` (itself relative to that root)
// combines with the current working directory to locate real output. If
// either is unavailable, extension files are always treated as absent.
func outputRoot(req *pb.GenerateRequest) string {
	out := req.GetSettings().GetCodegen().GetOut()
	if out == "" {
		return ""
	}
	if filepath.IsAbs(out) {
		return out
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, out)
}
