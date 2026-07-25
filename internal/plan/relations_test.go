package plan

import "testing"

// TestResolveTargetScan_NilQuery covers resolveTargetScan's nil-query
// guard. In the real pipeline buildRelationsPass2 only ever calls it when
// rr.LazyQuery/rr.EagerQuery is already non-nil, making this branch
// unreachable through the public config/generate entry points — covered
// directly here since resolveTargetScan is unexported.
func TestResolveTargetScan_NilQuery(t *testing.T) {
	scan, diags := resolveTargetScan(nil, nil, "SomeQuery", "path", "ctx", "Model", "Rel")
	if scan != nil || diags != nil {
		t.Fatalf("expected (nil, nil) for a nil query, got (%+v, %+v)", scan, diags)
	}
}
