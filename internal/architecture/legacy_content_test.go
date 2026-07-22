package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLegacyContentVerticalStaysRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)

	for _, relativePath := range []string{
		"internal/repository/post_repository.go",
		"internal/repository/post_category_repository.go",
		"internal/repository/banner_repository.go",
		"internal/repository/media_repository.go",
		"internal/service/post_service.go",
		"internal/service/post_category_service.go",
		"internal/service/banner_service.go",
		"internal/service/media_service.go",
		"internal/service/content_legacy_adapters.go",
		"internal/http/handlers/admin/content_compat.go",
		"internal/http/handlers/public/content_compat.go",
	} {
		if _, err := os.Stat(filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))); err == nil {
			t.Errorf("legacy Content file must not be restored: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect legacy Content file %s: %v", relativePath, err)
		}
	}

	assertTypesAbsent(t, filepath.Join(repositoryRoot, "internal", "repository"), map[string]struct{}{
		"PostRepository":         {},
		"PostCategoryRepository": {},
		"BannerRepository":       {},
		"MediaRepository":        {},
	})
	assertTypesAbsent(t, filepath.Join(repositoryRoot, "internal", "service"), map[string]struct{}{
		"PostService":         {},
		"PostCategoryService": {},
		"BannerService":       {},
		"MediaService":        {},
	})
	assertContainerFieldsAbsent(t, filepath.Join(repositoryRoot, "internal", "provider", "container.go"), map[string]struct{}{
		"PostRepo":            {},
		"PostCategoryRepo":    {},
		"BannerRepo":          {},
		"MediaRepo":           {},
		"PostService":         {},
		"PostCategoryService": {},
		"BannerService":       {},
		"MediaService":        {},
	})
	assertHandlerMethodsAbsent(t, filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin"), legacyAdminContentMethods())
	assertHandlerMethodsAbsent(t, filepath.Join(repositoryRoot, "internal", "http", "handlers", "public"), legacyPublicContentMethods())
	assertProductionImportsAbsent(t, filepath.Join(repositoryRoot, "internal", "service"), moduleImportPath+"/internal/modules/content")
}

func assertTypesAbsent(t *testing.T, directory string, forbidden map[string]struct{}) {
	t.Helper()
	forEachProductionGoFile(t, directory, func(path string, parsed *ast.File) {
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, blocked := forbidden[typeSpec.Name.Name]; blocked {
					t.Errorf("legacy Content type %s must not be declared in %s", typeSpec.Name.Name, path)
				}
			}
		}
	})
}

func assertContainerFieldsAbsent(t *testing.T, path string, forbidden map[string]struct{}) {
	t.Helper()
	parsed := parseProductionGoFile(t, path)
	for _, declaration := range parsed.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, spec := range general.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Container" {
				continue
			}
			container, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				t.Fatal("provider.Container is not a struct")
			}
			for _, field := range container.Fields.List {
				for _, name := range field.Names {
					if _, blocked := forbidden[name.Name]; blocked {
						t.Errorf("provider.Container must not expose legacy Content field %s", name.Name)
					}
				}
			}
			return
		}
	}
	t.Fatal("provider.Container declaration not found")
}

func assertHandlerMethodsAbsent(t *testing.T, directory string, forbidden map[string]struct{}) {
	t.Helper()
	forEachProductionGoFile(t, directory, func(path string, parsed *ast.File) {
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil {
				continue
			}
			if _, blocked := forbidden[function.Name.Name]; blocked {
				t.Errorf("legacy Content Handler method %s must not be declared in %s", function.Name.Name, path)
			}
		}
	})
}

func assertProductionImportsAbsent(t *testing.T, directory, forbidden string) {
	t.Helper()
	forEachProductionGoFile(t, directory, func(path string, parsed *ast.File) {
		for _, imported := range parsed.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatalf("unquote import in %s: %v", path, err)
			}
			if importMatches(importPath, forbidden) {
				t.Errorf("legacy service package must not import Content module in %s", path)
			}
		}
	})
}

func forEachProductionGoFile(t *testing.T, directory string, visit func(path string, parsed *ast.File)) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("read Go package %s: %v", directory, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		visit(path, parseProductionGoFile(t, path))
	}
}

func parseProductionGoFile(t *testing.T, path string) *ast.File {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return parsed
}

func legacyAdminContentMethods() map[string]struct{} {
	return map[string]struct{}{
		"GetAdminPosts":           {},
		"CreatePost":              {},
		"UpdatePost":              {},
		"DeletePost":              {},
		"GetAdminPostProductIDs":  {},
		"GetPostCategories":       {},
		"CreatePostCategory":      {},
		"UpdatePostCategory":      {},
		"DeletePostCategory":      {},
		"PatchPostCategoryStatus": {},
		"GetAdminBanners":         {},
		"GetAdminBanner":          {},
		"CreateBanner":            {},
		"UpdateBanner":            {},
		"DeleteBanner":            {},
		"GetAdminMedia":           {},
		"UpdateMedia":             {},
		"BatchDeleteMedia":        {},
		"DeleteMedia":             {},
	}
}

func legacyPublicContentMethods() map[string]struct{} {
	return map[string]struct{}{
		"GetPosts":          {},
		"GetPostBySlug":     {},
		"GetPublicBanners":  {},
		"GetPostCategories": {},
	}
}
