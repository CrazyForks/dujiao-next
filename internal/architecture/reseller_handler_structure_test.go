package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResellerHandlerIsSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	handlerDirectory := filepath.Join(repositoryRoot, "internal", "http", "handlers", "public")
	legacyPath := filepath.Join(handlerDirectory, "reseller.go")
	if _, err := os.Stat(legacyPath); err == nil {
		t.Fatalf("reseller.go must be replaced by responsibility-focused handler files")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat reseller.go: %v", err)
	}
	for _, migrated := range []string{
		"reseller_product.go",
		"reseller_finance.go",
		"reseller_order.go",
		"reseller_errors.go",
	} {
		if _, err := os.Stat(filepath.Join(handlerDirectory, migrated)); err == nil {
			t.Fatalf("%s must be migrated to transport/http/reseller", migrated)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", migrated, err)
		}
	}
}
