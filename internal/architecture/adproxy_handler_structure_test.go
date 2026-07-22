package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdProxyHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "adproxy")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "adproxy")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "types.go"), []string{
		"RenderSlot", "RenderItem", "RenderResponse",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"Service"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminHandler", "AdminService",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 3)

	legacy := filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_ad_proxy.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy ad proxy handler must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy ad proxy handler: %v", err)
	}
	legacyService := filepath.Join(repositoryRoot, "internal", "service", "ad_proxy_service.go")
	if _, err := os.Stat(legacyService); err == nil {
		t.Fatalf("legacy ad proxy service must stay removed: %s", legacyService)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy ad proxy service: %v", err)
	}
}
