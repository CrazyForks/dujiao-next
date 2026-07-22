package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalletQueryImplementationLivesInModule(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "wallet")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{
		"AccountListFilter", "TransactionListFilter", "RechargeListFilter", "Repository",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "query.go"), []string{"QueryService"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "query.go"), []string{
		"NewQueryService", "GetAccount", "ListTransactions", "ListRechargeOrdersAdmin",
		"ListUserRechargeOrders", "StatsUserRechargeOrders", "GetRechargeOrderByRechargeNo",
		"GetRechargeOrderByPaymentIDAndUser", "GetBalancesByUserIDs",
	})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "store.go"), []string{"Store"})
	assertFileDeclaresFunctions(t, filepath.Join(storeRoot, "store.go"), []string{"New", "WithTx"})
	assertDirectoryGoFileBudget(t, moduleRoot, 3)
	assertDirectoryGoFileBudget(t, storeRoot, 2)

	legacyImplementation := filepath.Join(repositoryRoot, "internal", "repository", "wallet_repository_impl.go")
	if _, err := os.Stat(legacyImplementation); err == nil {
		t.Fatalf("legacy wallet repository implementation must stay removed: %s", legacyImplementation)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", legacyImplementation, err)
	}
}
