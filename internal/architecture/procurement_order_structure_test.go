package architecture

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestProcurementOrderUsesBoundedContextLayout(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)

	legacyFiles := []string{
		"internal/repository/procurement_order_repository.go",
		"internal/http/handlers/admin/admin_procurement_order.go",
		"internal/service/procurement_order_service.go",
		"internal/service/procurement_order_create.go",
		"internal/service/procurement_order_submit.go",
		"internal/service/procurement_order_callback.go",
		"internal/service/procurement_order_poll.go",
		"internal/service/procurement_order_query.go",
		"internal/service/procurement_order_manual.go",
	}
	for _, relativePath := range legacyFiles {
		if _, err := os.Stat(filepath.Join(repositoryRoot, relativePath)); err == nil {
			t.Errorf("legacy procurement file must not return: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", relativePath, err)
		}
	}

	moduleDirectory := filepath.Join(repositoryRoot, "internal", "modules", "procurement")
	assertDirectoryGoFileBudget(t, moduleDirectory, 10)
	assertDirectoryGoFileBudget(t, filepath.Join(moduleDirectory, "store", "gormstore"), 3)
	assertDirectoryGoFileBudget(t, filepath.Join(repositoryRoot, "internal", "transport", "http", "procurement"), 3)

	expected := map[string][]string{
		"service.go": {"NewService"},
		"create.go":  {"CreateForOrder", "createProcurementForSingleOrder", "hasUpstreamItems"},
		"submit.go": {
			"SubmitToUpstream", "markProcurementError", "rejectProcurement",
			"rollbackLocalOrderOnProcurementFailure", "notifyProcurementFailure", "handleSubmitFailure",
			"isRetryableErrorCode", "parseRetryIntervals",
		},
		"callback.go": {"HandleUpstreamCallback", "createUpstreamFulfillment"},
		"poll.go":     {"PollUpstreamStatus", "requeuePoll", "SyncAcceptedOrders", "mapProcurementUpstreamStatus"},
		"query.go": {
			"GetByID", "GetByLocalOrderNo", "List", "StatsByStatus", "FillParentOrderNo", "fillParentOrderNos",
			"applyProcurementLocalRefundedAmountFallback", "shouldSyncUpstreamRefundStatus",
			"normalizeProcurementUpstreamStatus", "buildUpstreamRefundRecords", "parseUpstreamRefundRecordCreatedAt",
			"fillUpstreamRefundRecordsForProcurementOrder", "isPositiveUpstreamRefundAmount",
			"fillUpstreamRefundRecordsForProcurementOrders",
		},
		"manual.go": {"RetryManual", "CancelManual"},
	}
	for file, want := range expected {
		parsed := parseProductionGoFile(t, filepath.Join(moduleDirectory, file))
		got := declaredFunctionNames(parsed)
		sort.Strings(want)
		sort.Strings(got)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s function ownership mismatch\nwant: %v\ngot:  %v", file, want, got)
		}
	}
}
