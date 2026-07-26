package architecture

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestPaymentCallbackImplementationIsSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "modules", "payment", "application")
	expected := map[string][]string{
		"payment_service_callback.go": {
			"HandleCallback", "updateCallbackMeta", "applyPaymentUpdate",
			"mergeProviderPayload", "markOrderPaid",
		},
		"payment_service_callback_wallet.go": {
			"handleWalletRechargeCallback", "applyWalletRechargePaymentUpdate", "canApplyWalletRechargeCallback",
		},
		"payment_service_callback_dispatch.go": {
			"enqueueOrderPaidAsync", "enqueueProcurementAsync", "enqueueDownstreamCallbackAsync",
			"enqueueOrderPaidNotificationAsync", "enqueueWalletRechargeSuccessAsync",
			"enqueueOrderPaidBotNotifyAsync", "enqueueWalletRechargeBotNotifyAsync",
			"hasManualFulfillmentItems", "enqueueManualFulfillmentPendingAsync",
		},
		"payment_service_notification_payload.go": {
			"buildOrderNotificationPayload", "buildWalletRechargeNotificationPayload",
			"buildManualFulfillmentNotificationPayload", "notificationTemplateLocale",
			"resolveNotificationCustomer", "resolveUserNotificationIdentity",
			"notificationPaymentChannel", "notificationPayloadString",
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

func TestPaymentCallbackTypesLiveWithTheirResponsibilities(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "modules", "payment", "application")
	expectedOwner := map[string]string{
		"PaymentCallbackInput": "payment_service_callback.go",
	}
	actualOwner := make(map[string][]string, len(expectedOwner))
	files := []string{
		"payment_service_callback.go",
		"payment_service_callback_wallet.go",
		"payment_service_callback_dispatch.go",
		"payment_service_notification_payload.go",
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
