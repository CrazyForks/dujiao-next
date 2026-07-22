package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalletUserHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "wallet")

	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterUserRoutes",
		"RegisterAdminRoutes",
		"RegisterChannelRoutes",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_handler.go"), []string{
		"WalletService", "PaymentService", "UserReader", "SiteCurrencyReader", "UserHandler",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_handler.go"), []string{
		"NewUserHandler", "GetWallet", "GetTransactions", "GetPaymentChannels",
		"Recharge", "ListRecharges", "RechargeStats", "GetRecharge", "CaptureRechargePayment",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminWalletService", "AdminUserReader", "PaymentChannelReader", "PaymentReader",
		"AdminHandler", "AdminRechargeListFilter", "AdjustBalanceInput",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"NewAdminHandler", "GetUserWallet", "GetUserTransactions", "GetRecharges", "AdjustUserWallet",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "channel_handler.go"), []string{
		"ChannelUserProvisioner", "ChannelHandler",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "channel_handler.go"), []string{
		"NewChannelHandler", "GetWallet", "GetWalletTransactions", "CreateWalletRecharge",
	})
	assertDirectoryGoFileBudget(t, transportRoot, 5)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "wallet.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "wallet_admin.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "channel", "channel_wallet.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy wallet handler must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy wallet handler: %v", err)
		}
	}
	for _, legacy := range []string{"wallet_adapter.go", "giftcard_channel_user_adapter.go"} {
		if _, err := os.Stat(filepath.Join(repositoryRoot, "internal", "router", legacy)); err == nil {
			t.Fatalf("%s belongs in internal/wiring, not internal/router", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy router adapter %s: %v", legacy, err)
		}
	}
	assertDirectoryGoFileBudget(t, filepath.Join(repositoryRoot, "internal", "wiring", "wallet"), 4)
	assertDirectoryGoFileBudget(t, filepath.Join(repositoryRoot, "internal", "wiring", "channeluser"), 3)
}
