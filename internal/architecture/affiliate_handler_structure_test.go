package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAffiliateHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "affiliate")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "affiliate")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "types.go"), []string{
		"TrackClickInput", "WithdrawApplyInput", "Dashboard",
		"Stats", "AdminUserItem", "AdminProfileListFilter",
		"AdminCommissionListFilter", "AdminWithdrawListFilter",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterPublicRoutes", "RegisterUserRoutes",
		"RegisterAdminRoutes", "RegisterAdminFinanceRoutes",
		"RegisterChannelRoutes",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "handler.go"), []string{
		"Handler", "Service",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminHandler", "AdminService",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "channel_handler.go"), []string{
		"ChannelHandler", "ChannelUserProvisioner", "AffiliateSettings",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 3)
	assertDirectoryGoFileBudget(t, transportRoot, 4)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "affiliate.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_affiliate_manage.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "channel", "channel_affiliate.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy affiliate handler must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy affiliate handler: %v", err)
		}
	}
}
