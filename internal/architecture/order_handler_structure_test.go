package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOrderAdminHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "order")

	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterAdminRoutes", "RegisterAdminRefundRoutes", "RegisterAdminRefundWriteRoutes",
		"RegisterUserReadRoutes", "RegisterUserCancelRoute", "RegisterUserPreviewRoute", "RegisterUserCreateRoute",
		"RegisterUserCreateAndPayRoute", "RegisterUserPaymentChannelsRoute",
		"RegisterGuestReadRoutes", "RegisterGuestPreviewRoute", "RegisterGuestCreateRoute",
		"RegisterGuestCreateAndPayRoute",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminHandler", "OrderQuery", "UserDirectory", "CouponLookup", "PromotionLookup",
		"PaymentDirectory", "PaymentChannelDirectory", "OrderListFilter",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"NewAdminHandler", "AdminListOrders", "AdminGetOrder", "AdminUpdateOrderStatus",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_refund_handler.go"), []string{
		"AdminRefundHandler", "AdminRefundReader", "AdminRefundWriter", "AdminWalletRefunder",
		"OrderByIDLookup", "OrderStatusEmailEnqueuer", "AdminRefundListQuery", "AdminRefundItem",
		"AdminRefundToWalletInput", "AdminManualRefundInput",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "admin_refund_handler.go"), []string{
		"NewAdminRefundHandler", "GetAdminOrderRefunds", "GetAdminOrderRefund",
		"AdminRefundOrderToWallet", "AdminManualRefundOrder",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_handler.go"), []string{
		"UserHandler", "UserOrderQuery", "PaymentChannelPolicy", "RefundRecordDirectory", "UserLookup",
		"UserOrderListFilter", "AvailablePaymentChannelFilter", "OrderPaymentChannelsRequest",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_handler.go"), []string{
		"NewUserHandler", "ListOrders", "OrderStats", "GetOrderByOrderNo", "DownloadFulfillment",
		"GetOrderPaymentChannels", "CancelOrder",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "guest_handler.go"), []string{
		"GuestHandler", "GuestOrderQuery",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "guest_handler.go"), []string{
		"NewGuestHandler", "ListGuestOrders", "GetGuestOrderByOrderNo", "DownloadGuestFulfillment",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "preview_handler.go"), []string{
		"PreviewHandler", "OrderPreviewService", "OrderPreview", "CreateOrderInput", "CreateGuestOrderInput",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "preview_handler.go"), []string{
		"NewPreviewHandler", "PreviewOrder", "PreviewGuestOrder",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "create_handler.go"), []string{
		"CreateHandler", "OrderCreateService", "GuestCreateCaptcha", "OrderPaymentCreator",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "create_handler.go"), []string{
		"NewCreateHandler", "CreateOrder", "CreateGuestOrder", "CreateOrderAndPay", "CreateGuestOrderAndPay",
	})
	assertDirectoryGoFileBudget(t, transportRoot, 7)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "order_admin.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_order_refund.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "order.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "guest_order.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy admin order handler must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy admin order handler: %v", err)
		}
	}

	if _, err := os.Stat(filepath.Join(repositoryRoot, "internal", "router", "order_adapter.go")); err == nil {
		t.Fatal("order composition adapters belong in internal/wiring/order, not internal/router")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy order router adapter: %v", err)
	}
	wiringRoot := filepath.Join(repositoryRoot, "internal", "wiring", "order")
	for _, file := range []string{"wiring.go", "adapters.go"} {
		if _, err := os.Stat(filepath.Join(wiringRoot, file)); err != nil {
			t.Fatalf("order wiring file %s missing: %v", file, err)
		}
	}
	assertDirectoryGoFileBudget(t, wiringRoot, 4)
}
