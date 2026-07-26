package integration

import "testing"

func TestTransactionCommitRollbackAndPanicCleanup(t *testing.T) {
	runUserIntegrationE2E(t)
}
