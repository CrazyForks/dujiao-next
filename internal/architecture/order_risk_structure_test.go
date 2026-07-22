package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOrderRiskLivesInModule(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "orderrisk")
	serviceFile := filepath.Join(moduleRoot, "service.go")

	assertFileDeclaresTypes(t, serviceFile, []string{
		"Service", "SettingReader", "PendingOrderCounter", "CheckInput", "RiskRateLimitedError",
	})
	assertFileDeclaresFunctions(t, serviceFile, []string{"NewService", "GetRetryAfter"})
	assertDirectoryGoFileBudget(t, moduleRoot, 2)

	for _, relativePath := range []string{
		"internal/service/order_risk_control_service.go",
		"internal/service/order_risk_control_setting.go",
	} {
		path := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("legacy order risk file must stay removed: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", relativePath, err)
		}
	}
}
