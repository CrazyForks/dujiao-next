package architecture

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestResellerRepositoryImplementationIsSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	repositoryDirectory := filepath.Join(repositoryRoot, "internal", "repository")
	expected := map[string][]string{
		"reseller_repository.go": {"NewResellerRepository", "WithTx"},
		"reseller_profile_repository.go": {
			"CreateProfile", "GetProfileByID", "GetProfileByUserID", "UpdateProfile", "ListProfiles", "IsActiveRelatedAccount",
		},
		"reseller_domain_repository.go": {
			"UpsertDomain", "GetDomainByID", "GetDomainByIDForUpdate", "UpdateDomain", "FindDomainByHost",
			"FindActiveVerifiedDomain", "ListDomainsByResellerID", "ListDomains", "normalizeDomainForRepository",
		},
		"reseller_site_config_repository.go": {
			"UpsertSiteConfig", "GetSiteConfigByResellerID", "DeleteSiteConfigByResellerID", "ListSiteConfigs",
		},
		"reseller_pricing_read_repository.go": {
			"ListProductSettingsForPricing", "ListHiddenProductIDs", "uniqueUintSlice",
		},
		"reseller_order_snapshot_repository.go": {
			"CreateOrderSnapshot", "GetOrderSnapshotByOrderID", "applyResellerOrderSnapshotFilter",
			"ListOrderSnapshotsByReseller", "StatsOrderSnapshotsByReseller", "GetOrderSnapshotByResellerOrderNo",
			"buildResellerOrderSnapshotRows", "resellerOrderItemsFromParentOrChildren",
		},
		"reseller_ledger_repository.go": {
			"CreateLedgerEntryIfNotExists", "GetLedgerEntryByIdempotencyKey", "MarkDueLedgerEntriesAvailable",
			"ListDueLedgerScopes", "ListLedgerEntries", "SumLedgerAmount", "SumLedgerAmountByOrderAndType",
			"SumLedgerAmountGroupedByStatus", "ListAvailableLedgerEntriesForUpdate", "UpdateLedgerEntry",
			"BatchUpdateLedgerEntries", "BatchUpdateLedgerEntriesByWithdrawID",
		},
		"reseller_balance_repository.go": {
			"GetOrCreateBalanceAccountForUpdate", "ListBalanceAccounts", "UpdateBalanceAccount",
		},
		"reseller_withdraw_repository.go": {
			"CreateWithdrawRequest", "GetWithdrawRequestByID", "GetWithdrawRequestByIDForUpdate",
			"UpdateWithdrawRequest", "ListWithdrawRequests",
		},
		"reseller_admin_repository.go": {
			"ListAdminResellerLedgerEntries", "ListAdminResellerBalanceAccounts",
			"ListAdminResellerWithdrawRequests", "applyAdminResellerProfileFilters",
		},
	}

	for file, want := range expected {
		parsed := parseProductionGoFile(t, filepath.Join(repositoryDirectory, file))
		got := declaredFunctionNames(parsed)
		sort.Strings(want)
		sort.Strings(got)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s function ownership mismatch\nwant: %v\ngot:  %v", file, want, got)
		}
	}
}
