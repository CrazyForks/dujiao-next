package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReconciliationUsesBoundedContextLayout(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	for _, relativePath := range []string{
		"internal/service/reconciliation_service.go",
		"internal/repository/reconciliation_repository.go",
		"internal/http/handlers/admin/admin_reconciliation.go",
		"internal/http/handlers/admin/handler.go",
	} {
		path := filepath.Join(repositoryRoot, relativePath)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("legacy reconciliation file must stay removed: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", relativePath, err)
		}
	}

	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "reconciliation")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "reconciliation")
	assertDirectoryGoFileBudget(t, moduleRoot, 5)
	assertDirectoryGoFileBudget(t, storeRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 4)

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{
		"Service", "ServiceOptions", "RunInput", "JobListFilter", "JobRepository", "ItemRepository",
		"ProcurementReader", "ConnectionProvider", "Enqueuer", "NotificationEnqueuer",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService", "CreateAndEnqueue", "Execute"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "query.go"), []string{"GetJob", "ListJobs", "GetJobItems", "ResolveItem"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{"AdminHandler", "Service"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})
}
