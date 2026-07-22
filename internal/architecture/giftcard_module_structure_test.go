package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGiftCardImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "giftcard")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "giftcard")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{
		"ListFilter", "Repository", "UserDirectory", "CurrencyProvider",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "generate.go"), []string{"Generate"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "manage.go"), []string{
		"List", "Update", "Delete", "BatchUpdateStatus", "ResolveRedeemedUsers",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "export.go"), []string{"Export"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "store.go"), []string{"Store"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterAdminRoutes", "RegisterUserRoutes", "RegisterChannelRoutes",
	})

	assertDirectoryGoFileBudget(t, moduleRoot, 10)
	assertDirectoryGoFileBudget(t, storeRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 6)
}

func TestGiftCardLegacyRepositoryFileStaysRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	path := filepath.Join(repositoryRoot, "internal", "repository", "gift_card_repository.go")
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("legacy gift card repository must stay removed: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}
