package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSiteConnectionHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "siteconnection")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "siteconnection")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "types.go"), []string{
		"CreateInput", "UpdateInput", "ListFilter", "PingResult",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"Service", "Repository"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminHandler", "AdminService", "MarkupReapplier",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 3)
	assertDirectoryGoFileBudget(t, transportRoot, 3)

	legacy := filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_site_connection.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy site connection handler must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy site connection handler: %v", err)
	}
	legacyService := filepath.Join(repositoryRoot, "internal", "service", "site_connection_service.go")
	if _, err := os.Stat(legacyService); err == nil {
		t.Fatalf("legacy site connection service must stay removed: %s", legacyService)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy site connection service: %v", err)
	}
}
