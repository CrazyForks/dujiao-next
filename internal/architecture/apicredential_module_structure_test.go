package architecture

import (
	"path/filepath"
	"testing"
)

func TestApiCredentialImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "apicredential")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "apicredential")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{
		"ListFilter", "Repository",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{
		"NewService", "Apply", "Approve", "Reject", "SetActive", "SetActiveByUserID",
		"Regenerate", "RegenerateByUserID", "GetByUserID", "GetByID", "List", "Delete",
	})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "store.go"), []string{"Store"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterAdminRoutes", "RegisterUserRoutes",
	})

	assertDirectoryGoFileBudget(t, moduleRoot, 3)
	assertDirectoryGoFileBudget(t, storeRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 3)
}

func TestApiCredentialLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	patterns := []string{
		filepath.Join(repositoryRoot, "internal", "repository", "api_credential*.go"),
		filepath.Join(repositoryRoot, "internal", "service", "api_credential_service*.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_api_credential.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "api_credential.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		if len(matches) != 0 {
			t.Errorf("legacy API credential files must stay removed: %v", matches)
		}
	}
}
