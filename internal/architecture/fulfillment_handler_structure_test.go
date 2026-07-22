package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFulfillmentHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "fulfillment")

	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"AdminHandler", "ManualCreator", "AdminOrderReader", "CreateManualInput",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"NewAdminHandler", "AdminCreateFulfillment", "AdminDownloadFulfillment",
	})
	assertDirectoryGoFileBudget(t, transportRoot, 3)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "fulfillment_admin.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "cache.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy fulfillment/cache handler must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy handler: %v", err)
		}
	}
}
