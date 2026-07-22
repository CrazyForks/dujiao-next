package architecture

import (
	"path/filepath"
	"testing"
)

func TestCouponImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "coupon")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "coupon")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{
		"ListFilter", "UsageListFilter", "Repository", "UsageRepository",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "admin_service.go"), []string{
		"AdminService", "CreateCouponInput", "UpdateCouponInput",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"Service"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "store.go"), []string{"Store"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "usage_store.go"), []string{"UsageStore"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})

	assertDirectoryGoFileBudget(t, moduleRoot, 4)
	assertDirectoryGoFileBudget(t, storeRoot, 5)
	assertDirectoryGoFileBudget(t, transportRoot, 3)
}

func TestCouponLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	patterns := []string{
		filepath.Join(repositoryRoot, "internal", "service", "coupon*.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "coupon*.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_coupon*.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		if len(matches) != 0 {
			t.Errorf("legacy coupon files must stay removed: %v", matches)
		}
	}
}
