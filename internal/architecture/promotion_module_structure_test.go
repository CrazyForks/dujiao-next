package architecture

import (
	"path/filepath"
	"testing"
)

func TestPromotionImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "promotion")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "promotion")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{"ListFilter", "Repository"})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "admin_service.go"), []string{
		"AdminService", "CreatePromotionInput", "UpdatePromotionInput",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"Service"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "store.go"), []string{"Store"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{"AdminService", "AdminHandler"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})

	assertDirectoryGoFileBudget(t, moduleRoot, 4)
	assertDirectoryGoFileBudget(t, storeRoot, 3)
	assertDirectoryGoFileBudget(t, transportRoot, 3)
}

func TestPromotionLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	patterns := []string{
		filepath.Join(repositoryRoot, "internal", "service", "promotion*.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "promotion*.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_promotion*.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		if len(matches) != 0 {
			t.Errorf("legacy promotion files must stay removed: %v", matches)
		}
	}
}
