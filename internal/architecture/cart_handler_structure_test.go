package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCartHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "cart")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "cart")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{
		"Service", "Repository", "ProductReader", "SKUReader", "CurrencyReader", "ItemDetail", "UpsertItemInput",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterUserRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_handler.go"), []string{
		"UserHandler", "Service",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_handler.go"), []string{
		"NewUserHandler", "GetCart", "UpsertCartItem", "DeleteCartItem",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 1)
	assertDirectoryGoFileBudget(t, transportRoot, 4)

	legacy := filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "cart.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy cart handler must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy cart handler: %v", err)
	}
	for _, relativePath := range []string{
		"internal/service/cart_service.go",
		"internal/wiring/cart/wiring.go",
	} {
		path := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("legacy cart file must stay removed: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", relativePath, err)
		}
	}
}
