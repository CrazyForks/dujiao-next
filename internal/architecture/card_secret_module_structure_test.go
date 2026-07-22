package architecture

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestCardSecretUsesBoundedContextLayout(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	legacyFiles := []string{
		"internal/service/card_secret_core.go",
		"internal/service/card_secret_import.go",
		"internal/service/card_secret_manage.go",
		"internal/service/card_secret_export.go",
		"internal/service/card_secret_stats.go",
		"internal/service/card_secret_service_test.go",
		"internal/repository/card_secret_repository.go",
		"internal/repository/card_secret_batch_repository.go",
		"internal/http/handlers/admin/card_secret_admin.go",
	}
	for _, relativePath := range legacyFiles {
		if _, err := os.Stat(filepath.Join(repositoryRoot, relativePath)); err == nil {
			t.Errorf("legacy card secret file must not return: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", relativePath, err)
		}
	}

	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "cardsecret")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "cardsecret")
	assertDirectoryGoFileBudget(t, moduleRoot, 6)
	assertDirectoryGoFileBudget(t, storeRoot, 3)
	assertDirectoryGoFileBudget(t, transportRoot, 3)

	expected := map[string][]string{
		"service.go": {"NewService", "resolveCardSecretSKU", "normalizeCardSecretIDs"},
		"import.go": {
			"CreateCardSecretBatch", "ImportCardSecretCSV", "shouldDeduplicateCardSecrets",
			"normalizeSecrets", "parseCSVSecrets", "generateBatchNo",
		},
		"manage.go": {
			"ListCardSecrets", "buildRepositoryFilter", "hasListFilter",
			"BatchUpdateCardSecretStatus", "BatchDeleteCardSecrets", "UpdateCardSecret",
		},
		"export.go": {
			"ExportCardSecrets", "ExportAvailableCardSecrets", "normalizeCardSecretExportFormat",
			"buildCardSecretExportContent", "validateAutoCardSecretExportScope",
			"resolveExportTargetCardSecretIDs", "resolveBatchTargetCardSecretIDs",
		},
		"stats.go": {"GetStats", "ListBatches"},
	}
	for file, want := range expected {
		parsed := parseProductionGoFile(t, filepath.Join(moduleRoot, file))
		got := declaredFunctionNames(parsed)
		sort.Strings(want)
		sort.Strings(got)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s function ownership mismatch\nwant: %v\ngot:  %v", file, want, got)
		}
	}
}
