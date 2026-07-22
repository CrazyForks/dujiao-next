package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComplianceHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "compliance")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "compliance")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "types.go"), []string{
		"AcknowledgeRequest", "Status",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"Service", "SettingRepository"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminHandler", "AdminService",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 3)
	assertDirectoryGoFileBudget(t, transportRoot, 3)

	legacy := filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_compliance.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy compliance handler must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy compliance handler: %v", err)
	}
	legacyService := filepath.Join(repositoryRoot, "internal", "service", "compliance_service.go")
	if _, err := os.Stat(legacyService); err == nil {
		t.Fatalf("legacy compliance service must stay removed: %s", legacyService)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy compliance service: %v", err)
	}
}
