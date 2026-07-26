package architecture

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestRouterContainsRoutesAndMiddlewareOnly(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	routerRoot := filepath.Join(repositoryRoot, "internal", "router")

	adapters, err := filepath.Glob(filepath.Join(routerRoot, "*_adapter.go"))
	if err != nil {
		t.Fatalf("list router adapters: %v", err)
	}
	if len(adapters) != 0 {
		t.Fatalf("composition adapters belong in internal/bootstrap, not internal/router: %v", adapters)
	}

	entries, err := os.ReadDir(routerRoot)
	if err != nil {
		t.Fatalf("read router directory: %v", err)
	}
	productionFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		productionFiles = append(productionFiles, entry.Name())
	}
	sort.Strings(productionFiles)
	if len(productionFiles) > 13 {
		t.Fatalf("router production file budget exceeded: got %d files: %v", len(productionFiles), productionFiles)
	}
}

func TestBootstrapPackagesStayFocused(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	bootstrapRoot := filepath.Join(repositoryRoot, "internal", "bootstrap")
	entries, err := os.ReadDir(bootstrapRoot)
	if err != nil {
		t.Fatalf("read bootstrap directory: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			assertDirectoryGoFileBudget(t, filepath.Join(bootstrapRoot, entry.Name()), 4)
		})
	}
}
