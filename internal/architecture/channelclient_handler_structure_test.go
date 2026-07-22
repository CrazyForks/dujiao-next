package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChannelClientHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "channelclient")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "channelclient")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "types.go"), []string{"ClientDetail"})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"Service", "Repository"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminHandler", "AdminService",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 3)
	assertDirectoryGoFileBudget(t, transportRoot, 3)

	legacy := filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_channel_client.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy channel client handler must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy channel client handler: %v", err)
	}
	legacyService := filepath.Join(repositoryRoot, "internal", "service", "channel_client_service.go")
	if _, err := os.Stat(legacyService); err == nil {
		t.Fatalf("legacy channel client service must stay removed: %s", legacyService)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy channel client service: %v", err)
	}
}
