package architecture

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestOrderServiceTestsAreSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	legacyPath := filepath.Join(serviceDirectory, "order_service_test.go")
	if _, err := os.Stat(legacyPath); err == nil {
		t.Fatalf("order_service_test.go must be replaced by responsibility-focused test files")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat order_service_test.go: %v", err)
	}

	expected := map[string][]string{
		"order_service_helpers_test.go": {
			"TestMergeCreateOrderItems", "TestMergeCreateOrderItemsConflict",
			"TestApplyCouponDiscountToItems", "TestResolveManualFormSubmissionPreferOrderItemKey",
			"TestResolveManualFormSubmissionFallbackLegacyProductKey",
		},
		"order_service_cancel_test.go": {
			"TestCancelExpiredOrderExpiresPendingPayments", "setupCancelPaymentTestDB",
			"newPendingOrderForCancel", "newPaymentForOrder", "TestCancelOrderExpiresPendingPayments",
			"TestUpdateOrderStatusAdminCancelExpiresPendingPaymentsSingleOrder",
			"TestCancelExpiredOrderExpiresPaymentsForParentAndChildren",
		},
		"order_service_status_test.go": {
			"TestCalcParentStatus", "TestCalcParentStatusAllRefunded", "TestCalcParentStatusPartiallyRefunded",
			"TestExpectedRefundStatus", "TestResolvedParentStatusPrefersOwnRefund",
			"TestIsTransitionAllowedRefunded", "TestUpdateOrderStatusParentToPartiallyRefundedSyncsChildren",
			"TestCanCompleteParentOrder", "TestCanCompleteParentOrderRejectInvalidStatus",
			"TestCanCompleteParentOrderRejectInvalidChild",
		},
		"order_service_pricing_test.go": {
			"assertBuildOrderResultRejectsPurchaseQuantity", "TestBuildOrderResultRejectsZeroPromotionPrice",
			"TestPreviewOrderAppliesMemberDiscountForManualProductBeforeFormCompleted",
			"TestBuildOrderResultStacksPromotionAndMemberDiscount",
			"TestBuildOrderResultRejectsProductMaxPurchaseQuantityExceeded",
			"TestBuildOrderResultRejectsProductMinPurchaseQuantityNotMet",
			"TestBuildOrderResultOriginalAmountBeforePromotion",
			"TestBuildOrderResultRejectsZeroTotalAmountAfterCoupon",
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

func TestOrderServiceTestFixtureLivesWithPricingTests(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	files := []string{
		"order_service_helpers_test.go",
		"order_service_cancel_test.go",
		"order_service_status_test.go",
		"order_service_pricing_test.go",
	}

	owners := make(map[string]string)
	for _, file := range files {
		parsed := parseProductionGoFile(t, filepath.Join(serviceDirectory, file))
		for _, typeName := range declaredTypeNames(parsed) {
			if previous, exists := owners[typeName]; exists {
				t.Fatalf("type %s declared in both %s and %s", typeName, previous, file)
			}
			owners[typeName] = file
		}
	}
	if got := owners["orderPurchaseQuantityLimitFixture"]; got != "order_service_pricing_test.go" {
		t.Fatalf("orderPurchaseQuantityLimitFixture must live in pricing tests, got %q", got)
	}
}
