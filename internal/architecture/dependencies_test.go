package architecture

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const moduleImportPath = "github.com/dujiao-next"

type importViolation struct {
	file       string
	importPath string
	reason     string
}

func TestDependencyRules(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	targets := []string{
		filepath.Join(repositoryRoot, "internal", "modules"),
		filepath.Join(repositoryRoot, "internal", "transport", "http"),
	}

	var violations []importViolation
	for _, target := range targets {
		found, err := inspectImports(repositoryRoot, target)
		if err != nil {
			t.Fatalf("inspect imports below %s: %v", target, err)
		}
		violations = append(violations, found...)
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].file == violations[j].file {
			return violations[i].importPath < violations[j].importPath
		}
		return violations[i].file < violations[j].file
	})

	for _, violation := range violations {
		t.Errorf("%s imports %q: %s", violation.file, violation.importPath, violation.reason)
	}
}

func TestValidateImportRules(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		importPath    string
		wantViolation bool
	}{
		{
			name:          "content core cannot import gorm",
			file:          "internal/modules/content/post_service.go",
			importPath:    "gorm.io/gorm",
			wantViolation: true,
		},
		{
			name:          "content core cannot import gin",
			file:          "internal/modules/content/post_service.go",
			importPath:    "github.com/gin-gonic/gin",
			wantViolation: true,
		},
		{
			name:          "content core cannot import legacy repository",
			file:          "internal/modules/content/post_service.go",
			importPath:    moduleImportPath + "/internal/repository",
			wantViolation: true,
		},
		{
			name:          "gorm store can import gorm",
			file:          "internal/modules/content/store/gormstore/post_store.go",
			importPath:    "gorm.io/gorm",
			wantViolation: false,
		},
		{
			name:          "gorm store cannot import legacy service",
			file:          "internal/modules/content/store/gormstore/post_store.go",
			importPath:    moduleImportPath + "/internal/service",
			wantViolation: true,
		},
		{
			name:          "content core can temporarily import shared models",
			file:          "internal/modules/content/post_service.go",
			importPath:    moduleImportPath + "/internal/models",
			wantViolation: false,
		},
		{
			name:          "content transport can import gin",
			file:          "internal/transport/http/content/public_handler.go",
			importPath:    "github.com/gin-gonic/gin",
			wantViolation: false,
		},
		{
			name:          "content transport cannot import repository",
			file:          "internal/transport/http/content/public_handler.go",
			importPath:    moduleImportPath + "/internal/repository",
			wantViolation: true,
		},
		{
			name:          "content transport cannot import service",
			file:          "internal/transport/http/content/public_handler.go",
			importPath:    moduleImportPath + "/internal/service",
			wantViolation: true,
		},
		{
			name:          "content transport cannot import provider container",
			file:          "internal/transport/http/content/public_handler.go",
			importPath:    moduleImportPath + "/internal/provider",
			wantViolation: true,
		},
		{
			name:          "dashboard core cannot import legacy service",
			file:          "internal/modules/dashboard/service.go",
			importPath:    moduleImportPath + "/internal/service",
			wantViolation: true,
		},
		{
			name:          "dashboard gorm store can import gorm",
			file:          "internal/modules/dashboard/store/gormstore/store.go",
			importPath:    "gorm.io/gorm",
			wantViolation: false,
		},
		{
			name:          "nested catalog product gorm store can import gorm",
			file:          "internal/modules/catalog/product/store/gormstore/product_store.go",
			importPath:    "gorm.io/gorm",
			wantViolation: false,
		},
		{
			name:          "dashboard transport cannot import legacy repository",
			file:          "internal/transport/http/dashboard/admin_handler.go",
			importPath:    moduleImportPath + "/internal/repository",
			wantViolation: true,
		},
		{
			name:          "content black box integration test can import gorm",
			file:          "internal/transport/http/content/admin_integration_test.go",
			importPath:    "gorm.io/gorm",
			wantViolation: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reason := validateImport(test.file, test.importPath)
			if test.wantViolation && reason == "" {
				t.Fatalf("expected %s importing %q to violate a dependency rule", test.file, test.importPath)
			}
			if !test.wantViolation && reason != "" {
				t.Fatalf("expected %s importing %q to be allowed, got: %s", test.file, test.importPath, reason)
			}
		})
	}
}

func TestInspectImportsRejectsForbiddenDependency(t *testing.T) {
	repositoryRoot := t.TempDir()
	contentRoot := filepath.Join(repositoryRoot, "internal", "modules", "content")
	if err := os.MkdirAll(contentRoot, 0o755); err != nil {
		t.Fatalf("create synthetic content module: %v", err)
	}

	source := []byte("package content\n\nimport _ \"gorm.io/gorm\"\n")
	if err := os.WriteFile(filepath.Join(contentRoot, "post_service.go"), source, 0o600); err != nil {
		t.Fatalf("write synthetic content source: %v", err)
	}

	violations, err := inspectImports(repositoryRoot, contentRoot)
	if err != nil {
		t.Fatalf("inspect synthetic content module: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected one dependency violation, got %d: %#v", len(violations), violations)
	}
	if violations[0].importPath != "gorm.io/gorm" {
		t.Fatalf("expected GORM violation, got %#v", violations[0])
	}
}

func findRepositoryRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test filename")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func inspectImports(repositoryRoot, target string) ([]importViolation, error) {
	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var violations []importViolation
	err := filepath.WalkDir(target, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		relativePath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return fmt.Errorf("resolve relative path for %s: %w", path, err)
		}
		relativePath = filepath.ToSlash(relativePath)

		for _, imported := range parsed.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return fmt.Errorf("unquote import %s in %s: %w", imported.Path.Value, path, err)
			}
			if reason := validateImport(relativePath, importPath); reason != "" {
				violations = append(violations, importViolation{
					file:       relativePath,
					importPath: importPath,
					reason:     reason,
				})
			}
		}

		return nil
	})

	return violations, err
}

func validateImport(file, importPath string) string {
	file = filepath.ToSlash(file)

	if pathWithin(file, "internal/modules") {
		if forbiddenLegacyImport(importPath) {
			return "domain modules must not depend on legacy service, repository, HTTP, router, or provider packages"
		}
		if importMatches(importPath, "github.com/gin-gonic/gin") {
			return "HTTP transport belongs outside domain modules"
		}
		if strings.HasPrefix(importPath, "gorm.io/") && !isGormStore(file) {
			return "only a module's store/gormstore adapter may import GORM"
		}
	}

	if pathWithin(file, "internal/transport/http") {
		for _, forbidden := range []string{
			moduleImportPath + "/internal/repository",
			moduleImportPath + "/internal/service",
			moduleImportPath + "/internal/provider",
		} {
			if importMatches(importPath, forbidden) {
				return "HTTP transport must depend on narrow use-case interfaces, not legacy implementations or the provider container"
			}
		}
		if strings.HasPrefix(importPath, "gorm.io/") && !strings.HasSuffix(file, "_integration_test.go") {
			return "HTTP transport must not depend on GORM"
		}
	}

	return ""
}

func forbiddenLegacyImport(importPath string) bool {
	for _, forbidden := range []string{
		moduleImportPath + "/internal/service",
		moduleImportPath + "/internal/repository",
		moduleImportPath + "/internal/http",
		moduleImportPath + "/internal/router",
		moduleImportPath + "/internal/provider",
	} {
		if importMatches(importPath, forbidden) {
			return true
		}
	}
	return false
}

func importMatches(importPath, forbidden string) bool {
	return importPath == forbidden || strings.HasPrefix(importPath, forbidden+"/")
}

func pathWithin(file, directory string) bool {
	return file == directory || strings.HasPrefix(file, directory+"/")
}

func isGormStore(file string) bool {
	parts := strings.Split(filepath.ToSlash(file), "/")
	if len(parts) < 6 || parts[0] != "internal" || parts[1] != "modules" {
		return false
	}
	for index := 2; index+1 < len(parts); index++ {
		if parts[index] == "store" && parts[index+1] == "gormstore" {
			return true
		}
	}
	return false
}
