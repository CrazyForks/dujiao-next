package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDownstreamCallbackLivesInModule(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "downstreamcallback")
	serviceFile := filepath.Join(moduleRoot, "service.go")

	assertFileDeclaresTypes(t, serviceFile, []string{
		"Service", "RefRepository", "OrderReader", "CredentialReader", "CallbackQueue",
	})
	assertFileDeclaresFunctions(t, serviceFile, []string{"NewService"})
	assertDirectoryGoFileBudget(t, moduleRoot, 1)

	legacy := filepath.Join(repositoryRoot, "internal", "service", "downstream_callback_service.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy downstream callback service must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy downstream callback service: %v", err)
	}
}
