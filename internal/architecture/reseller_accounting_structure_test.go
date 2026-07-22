package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResellerAccountingIsSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	legacyPath := filepath.Join(serviceDirectory, "reseller_accounting.go")
	if _, err := os.Stat(legacyPath); err == nil {
		t.Fatalf("reseller_accounting.go must be replaced by responsibility-focused service files")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat reseller_accounting.go: %v", err)
	}

	expectedOwner := map[string]string{
		"NewResellerAccountingService": "reseller_accounting_core.go",
		"getResellerProfileByUserID":   "reseller_accounting_core.go",
		"requireActiveResellerProfile": "reseller_accounting_core.go",
		"resellerWithdrawAvailability": "reseller_accounting_core.go",
		"refreshBalanceAccountTx":      "reseller_accounting_core.go",
		"GetUserFinanceDashboard":      "reseller_accounting_query.go",
		"ListUserBalanceAccounts":      "reseller_accounting_query.go",
		"ListUserLedgerEntries":        "reseller_accounting_query.go",
		"ListUserWithdrawRequests":     "reseller_accounting_query.go",
		"ListAdminLedgerEntries":       "reseller_accounting_query.go",
		"ListAdminBalanceAccounts":     "reseller_accounting_query.go",
		"ListAdminWithdrawRequests":    "reseller_accounting_query.go",
		"PostOrderProfitTx":            "reseller_accounting_profit.go",
		"ConfirmDueLedgerEntries":      "reseller_accounting_profit.go",
		"HandleRefundDeductTx":         "reseller_accounting_refund.go",
		"ApplyUserWithdraw":            "reseller_accounting_withdraw.go",
		"ApplyWithdraw":                "reseller_accounting_withdraw.go",
		"ReviewWithdraw":               "reseller_accounting_withdraw.go",
	}
	expectedTypeOwner := map[string]string{
		"ResellerAccountingOptions":             "reseller_accounting_core.go",
		"ResellerAccountingService":             "reseller_accounting_core.go",
		"ResellerAdminLedgerListFilter":         "reseller_accounting_query.go",
		"ResellerAdminBalanceAccountListFilter": "reseller_accounting_query.go",
		"ResellerAdminWithdrawListFilter":       "reseller_accounting_query.go",
		"ResellerUserFinanceDashboard":          "reseller_accounting_query.go",
		"ResellerUserLedgerListFilter":          "reseller_accounting_query.go",
		"ResellerUserBalanceAccountListFilter":  "reseller_accounting_query.go",
		"ResellerUserWithdrawListFilter":        "reseller_accounting_query.go",
		"ResellerWithdrawApplyInput":            "reseller_accounting_withdraw.go",
	}

	files := []string{
		"reseller_accounting_core.go",
		"reseller_accounting_query.go",
		"reseller_accounting_profit.go",
		"reseller_accounting_refund.go",
		"reseller_accounting_withdraw.go",
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

	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "reseller")
	for _, required := range []string{
		"accounting_profit.go",
		"accounting_refund.go",
		"accounting_balance.go",
		"accounting_withdraw.go",
		"accounting_query.go",
	} {
		if _, err := os.Stat(filepath.Join(moduleRoot, required)); err != nil {
			t.Fatalf("modules/reseller/%s must exist after accounting usecase migration: %v", required, err)
		}
	}
}
