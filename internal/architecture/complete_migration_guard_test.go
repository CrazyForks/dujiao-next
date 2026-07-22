package architecture

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// legacyRootGoFileBudgets freeze the remaining horizontal architecture at the
// start of the complete migration. Every entry is a one-way budget: deleting
// or moving files is allowed, adding files below these roots is not. The map is
// removed together with the roots when the migration is complete.
var legacyRootGoFileBudgets = map[string]int{
	"internal/dto":            24,
	"internal/http":           17,
	"internal/integration":    25,
	"internal/models":         59,
	"internal/payment":        53,
	"internal/persistence":    3,
	"internal/provider":       10,
	"internal/repository":     54,
	"internal/router":         23,
	"internal/service":        141,
	"internal/transport/http": 184,
	"internal/wiring":         51,
	"internal/worker":         10,
}

// These are the only production compatibility shims that existed at the
// migration baseline. They may be deleted, but no replacement shim may be
// introduced. The final migration gate requires this set to become empty.
var baselineCompatibilityFiles = map[string]struct{}{
	"internal/repository/card_secret_compat.go":      {},
	"internal/repository/product_compat.go":          {},
	"internal/repository/product_mapping_compat.go":  {},
	"internal/service/product_application_compat.go": {},
}

type packageFileBudget struct {
	production int
	total      int
}

// Existing oversized packages are frozen at their baseline size. All new or
// already-focused packages use the default limit below. Entries disappear as
// their packages are split into bounded-context leaf packages.
var transitionalPackageFileBudgets = map[string]packageFileBudget{
	"internal/architecture":            {production: 0, total: 54},
	"internal/dto":                     {production: 15, total: 24},
	"internal/models":                  {production: 54, total: 59},
	"internal/modules/reseller":        {production: 19, total: 26},
	"internal/payment/provider":        {production: 14, total: 25},
	"internal/repository":              {production: 36, total: 54},
	"internal/router":                  {production: 13, total: 23},
	"internal/service":                 {production: 81, total: 141},
	"internal/transport/http/reseller": {production: 13, total: 20},
}

// completedMigrationPaths are deleted compatibility-free entry points. Once a
// bounded context reaches this list, recreating its former horizontal package
// is an architecture regression rather than an allowed transitional change.
var completedMigrationPaths = []string{
	"internal/http/handlers/shared/captcha_payload.go",
	"internal/models/setting.go",
	"internal/repository/setting_repository.go",
	"internal/service/setting_service.go",
	"internal/transport/http/captcha",
	"internal/transport/http/settings",
	"internal/wiring/captcha",
	"internal/wiring/settings",
}

func TestCompletedMigrationPathsStayDeleted(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	for _, relativePath := range completedMigrationPaths {
		t.Run(strings.ReplaceAll(relativePath, "/", "_"), func(t *testing.T) {
			absolutePath := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
			if _, err := os.Stat(absolutePath); err == nil {
				t.Fatalf("completed migration path was recreated: %s", relativePath)
			} else if !os.IsNotExist(err) {
				t.Fatalf("stat completed migration path %s: %v", relativePath, err)
			}
		})
	}
}

func TestSettingsModuleRootContainsNoGoFiles(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	settingsRoot := filepath.Join(repositoryRoot, "internal", "modules", "settings")
	production, total := countDirectGoFiles(t, settingsRoot)
	if production != 0 || total != 0 {
		t.Fatalf("settings module root must remain structural only, got production=%d total=%d", production, total)
	}
}

func TestLegacyHorizontalRootsCanOnlyShrink(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)

	paths := make([]string, 0, len(legacyRootGoFileBudgets))
	for relativePath := range legacyRootGoFileBudgets {
		paths = append(paths, relativePath)
	}
	sort.Strings(paths)

	for _, relativePath := range paths {
		t.Run(strings.TrimPrefix(relativePath, "internal/"), func(t *testing.T) {
			absolutePath := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
			count := countGoFilesRecursively(t, absolutePath)
			if count > legacyRootGoFileBudgets[relativePath] {
				t.Fatalf(
					"legacy root %s grew from its migration baseline of %d Go files to %d; move new code into a bounded context",
					relativePath,
					legacyRootGoFileBudgets[relativePath],
					count,
				)
			}
		})
	}
}

func TestNoNewCompatibilityOrLegacyProductionFiles(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	internalRoot := filepath.Join(repositoryRoot, "internal")

	err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		name := strings.ToLower(entry.Name())
		if !strings.Contains(name, "compat") && !strings.Contains(name, "legacy") && !strings.Contains(name, "facade") {
			return nil
		}

		relativePath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)
		if _, allowedDuringMigration := baselineCompatibilityFiles[relativePath]; !allowedDuringMigration {
			t.Errorf("new compatibility or legacy production file is forbidden: %s", relativePath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect production Go files: %v", err)
	}
}

func TestGoPackagesStayWithinFileBudgets(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	internalRoot := filepath.Join(repositoryRoot, "internal")

	err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() || entry.Name() == "testdata" {
			return nil
		}

		production, total := countDirectGoFiles(t, path)
		if total == 0 {
			return nil
		}

		relativePath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)

		budget := packageFileBudget{production: 12, total: 20}
		if transitional, ok := transitionalPackageFileBudgets[relativePath]; ok {
			budget = transitional
		}

		if production > budget.production || total > budget.total {
			t.Errorf(
				"Go package %s exceeds its migration budget: production=%d/%d total=%d/%d",
				relativePath,
				production,
				budget.production,
				total,
				budget.total,
			)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect Go package budgets: %v", err)
	}
}

func countGoFilesRecursively(t *testing.T, root string) int {
	t.Helper()
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("stat %s: %v", root, err)
	}

	count := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && filepath.Ext(path) == ".go" {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("count Go files below %s: %v", root, err)
	}
	return count
}

func countDirectGoFiles(t *testing.T, directory string) (production int, total int) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read %s: %v", directory, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		total++
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			production++
		}
	}
	return production, total
}
