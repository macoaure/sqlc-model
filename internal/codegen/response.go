package codegen

import (
	"fmt"
	"sort"

	"github.com/macoaure/sqlc-gen-richmodel/internal/diagnostics"
	"github.com/macoaure/sqlc-gen-richmodel/internal/plan"

	pb "github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// Render renders every generator-owned file for p, plus each model's
// developer-owned extension file when it doesn't already exist in
// outputRoot (FR-015). The returned file list is sorted by path (FR-018,
// SC-002); a nil result means at least one error-severity diagnostic
// occurred and no output should be emitted for this run (FR-017).
func Render(p *plan.Plan, outputRoot string) ([]*pb.File, []diagnostics.Diagnostic) {
	var diags []diagnostics.Diagnostic
	files := make(map[string][]byte)

	add := func(path string, contents []byte, fdiags []diagnostics.Diagnostic) {
		diags = append(diags, fdiags...)
		if len(fdiags) == 0 {
			files[path] = contents
		}
	}

	for _, ctx := range p.Contexts {
		sessionOut, sdiags := RenderSession(ctx)
		add(fmt.Sprintf("%s/session_gen.go", ctx.Directory), sessionOut, sdiags)

		fieldsOut, fdiags := RenderFields(ctx)
		add(fmt.Sprintf("%s/fields_gen.go", ctx.Directory), fieldsOut, fdiags)

		for _, m := range ctx.Models {
			recordOut, rdiags := RenderRecord(ctx, m)
			add(fmt.Sprintf("%s/%s_record_gen.go", ctx.Directory, fileStem(m.Row)), recordOut, rdiags)

			storeOut, stdiags := RenderStore(ctx, m)
			add(fmt.Sprintf("%s/%s_store_gen.go", ctx.Directory, fileStem(m.Row)), storeOut, stdiags)

			collectionOut, cdiags := RenderCollection(ctx, m)
			add(fmt.Sprintf("%s/%s_collection_gen.go", ctx.Directory, fileStem(m.Row)), collectionOut, cdiags)

			modelOut, mdiags := RenderModel(ctx, m)
			add(fmt.Sprintf("%s/%s_gen.go", ctx.Directory, fileStem(m.Row)), modelOut, mdiags)

			for _, rel := range m.Relations {
				relOut, reldiags := RenderRelation(ctx, m, rel)
				add(fmt.Sprintf("%s/%s_%s_relation_gen.go", ctx.Directory, fileStem(m.Row), fileStem(rel.Name)), relOut, reldiags)
			}
			if hasEagerRelation(m) {
				eagerOut, eagerdiags := RenderEager(ctx, m)
				add(fmt.Sprintf("%s/%s_eager_gen.go", ctx.Directory, fileStem(m.Row)), eagerOut, eagerdiags)
			}

			if ef, emit := PlanExtensionFile(outputRoot, ctx, m); emit {
				files[ef.Path] = ef.Contents
			}
		}
	}

	if diagnostics.HasError(diags) {
		return nil, diags
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	out := make([]*pb.File, 0, len(paths))
	for _, path := range paths {
		out = append(out, &pb.File{Name: path, Contents: files[path]})
	}
	return out, diags
}
