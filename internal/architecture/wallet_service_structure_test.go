package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalletServiceIsSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	legacyPath := filepath.Join(serviceDirectory, "wallet_service.go")
	if _, err := os.Stat(legacyPath); err == nil {
		t.Fatalf("wallet_service.go must be replaced by responsibility-focused service files")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat wallet_service.go: %v", err)
	}

	expectedOwner := map[string]string{
		"NewWalletService":                   "wallet_core.go",
		"SetResellerAccountingService":       "wallet_core.go",
		"normalizeWalletCurrency":            "wallet_core.go",
		"cleanWalletRemark":                  "wallet_core.go",
		"buildOrderWalletReference":          "wallet_core.go",
		"buildWalletReference":               "wallet_core.go",
		"queries":                            "wallet_query.go",
		"GetAccount":                         "wallet_query.go",
		"ListTransactions":                   "wallet_query.go",
		"ListRechargeOrdersAdmin":            "wallet_query.go",
		"ListUserRechargeOrders":             "wallet_query.go",
		"StatsUserRechargeOrders":            "wallet_query.go",
		"GetRechargeOrderByRechargeNo":       "wallet_query.go",
		"GetRechargeOrderByPaymentIDAndUser": "wallet_query.go",
		"GetBalancesByUserIDs":               "wallet_query.go",
		"Recharge":                           "wallet_admin.go",
		"AdminAdjustBalance":                 "wallet_admin.go",
		"AdminRefundToWallet":                "wallet_admin.go",
		"ApplyOrderBalance":                  "wallet_order.go",
		"ReleaseOrderBalance":                "wallet_order.go",
		"ApplyRechargePayment":               "wallet_recharge.go",
		"CreditInTx":                         "wallet_credit.go",
		"changeBalance":                      "wallet_credit.go",
		"getOrCreateAccount":                 "wallet_credit.go",
		"ensureAccountForUpdate":             "wallet_credit.go",
	}
	expectedTypeOwner := map[string]string{
		"WalletService":            "wallet_core.go",
		"WalletRechargeInput":      "wallet_admin.go",
		"WalletAdjustInput":        "wallet_admin.go",
		"AdminRefundToWalletInput": "wallet_admin.go",
		"WalletCreditInput":        "wallet_credit.go",
	}

	files := []string{
		"wallet_core.go",
		"wallet_query.go",
		"wallet_admin.go",
		"wallet_order.go",
		"wallet_recharge.go",
		"wallet_credit.go",
	}
	actualOwners := make(map[string][]string, len(expectedOwner))
	actualTypeOwners := make(map[string][]string, len(expectedTypeOwner))
	for _, file := range files {
		parsed := parseProductionGoFile(t, filepath.Join(serviceDirectory, file))
		for _, function := range declaredFunctionNames(parsed) {
			if _, tracked := expectedOwner[function]; tracked {
				actualOwners[function] = append(actualOwners[function], file)
			}
		}
		for _, typeName := range declaredTypeNames(parsed) {
			if _, tracked := expectedTypeOwner[typeName]; tracked {
				actualTypeOwners[typeName] = append(actualTypeOwners[typeName], file)
			}
		}
	}

	for function, wantFile := range expectedOwner {
		gotFiles := actualOwners[function]
		if len(gotFiles) != 1 || gotFiles[0] != wantFile {
			t.Errorf("%s ownership mismatch: want [%s], got %v", function, wantFile, gotFiles)
		}
	}
	for typeName, wantFile := range expectedTypeOwner {
		gotFiles := actualTypeOwners[typeName]
		if len(gotFiles) != 1 || gotFiles[0] != wantFile {
			t.Errorf("%s ownership mismatch: want [%s], got %v", typeName, wantFile, gotFiles)
		}
	}
}
