package integration

import "testing"

func TestRelationCachingEagerLoadingAndInverseHydration(t *testing.T) {
	runUserIntegrationE2E(t)
}
