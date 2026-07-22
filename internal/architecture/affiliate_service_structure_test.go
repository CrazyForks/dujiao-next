package architecture

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestAffiliateServiceImplementationIsSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	expected := map[string][]string{
		"affiliate_service.go": {"NewAffiliateService"},
		"affiliate_profile.go": {
			"UpdateAffiliateProfileStatus", "BatchUpdateAffiliateProfileStatus", "OpenAffiliate",
			"normalizeAffiliateProfileIDs", "generateAffiliateCode", "isUniqueViolation",
		},
		"affiliate_attribution.go": {"ResolveOrderAffiliateSnapshot", "TrackClick"},
		"affiliate_commission.go": {
			"HandleOrderPaid", "ConfirmDueCommissions", "HandleOrderCanceled", "HandleOrderRefundedTx",
			"resolveAffiliateProfileForOrder", "calculateCommissionBaseAmount",
			"collectAffiliateProductIDs", "buildSplitCommissionType",
		},
		"affiliate_withdraw.go": {"ApplyWithdraw", "ReviewWithdraw", "containsWithdrawChannel"},
		"affiliate_query.go": {
			"GetUserDashboard", "ListUserCommissions", "ListUserWithdraws", "ListAdminUsers",
			"ListAdminCommissions", "ListAdminWithdraws", "buildProfileStats", "calcAffiliateConversion",
		},
	}

	for file, want := range expected {
		parsed := parseProductionGoFile(t, filepath.Join(serviceDirectory, file))
		got := declaredFunctionNames(parsed)
		sort.Strings(want)
		sort.Strings(got)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s function ownership mismatch\nwant: %v\ngot:  %v", file, want, got)
		}
	}
}

func TestAffiliateServiceTypesLiveWithTheirResponsibilities(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	expectedOwner := map[string]string{
		"AffiliateService":                   "affiliate_service.go",
		"AffiliateTrackClickInput":           "affiliate_attribution.go",
		"AffiliateWithdrawApplyInput":        "affiliate_withdraw.go",
		"AffiliateDashboard":                 "affiliate_query.go",
		"AffiliateStats":                     "affiliate_query.go",
		"AffiliateAdminUserItem":             "affiliate_query.go",
		"AffiliateAdminCommissionListFilter": "affiliate_query.go",
		"AffiliateAdminWithdrawListFilter":   "affiliate_query.go",
	}
	actualOwner := make(map[string][]string, len(expectedOwner))
	files := []string{
		"affiliate_service.go", "affiliate_profile.go", "affiliate_attribution.go",
		"affiliate_commission.go", "affiliate_withdraw.go", "affiliate_query.go",
	}

	for _, file := range files {
		parsed := parseProductionGoFile(t, filepath.Join(serviceDirectory, file))
		for _, typeName := range declaredTypeNames(parsed) {
			if _, tracked := expectedOwner[typeName]; tracked {
				actualOwner[typeName] = append(actualOwner[typeName], file)
			}
		}
	}

	for typeName, wantFile := range expectedOwner {
		gotFiles := actualOwner[typeName]
		if len(gotFiles) != 1 || gotFiles[0] != wantFile {
			t.Errorf("%s ownership mismatch: want [%s], got %v", typeName, wantFile, gotFiles)
		}
	}
}
