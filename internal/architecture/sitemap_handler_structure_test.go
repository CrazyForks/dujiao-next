package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSitemapHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "sitemap")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "sitemap")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{
		"Service", "SitemapPost", "PublishedPostReader", "PublishedPostReaderFunc",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "handler.go"), []string{
		"Handler", "Generator", "SiteBrandReader",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 1)
	assertDirectoryGoFileBudget(t, transportRoot, 4)

	legacy := filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "sitemap.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy public sitemap handler must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy sitemap handler: %v", err)
	}
	legacyService := filepath.Join(repositoryRoot, "internal", "service", "sitemap_service.go")
	if _, err := os.Stat(legacyService); err == nil {
		t.Fatalf("legacy sitemap service must stay removed: %s", legacyService)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy sitemap service: %v", err)
	}
}
