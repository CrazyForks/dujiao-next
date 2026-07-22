package architecture

import (
	"path/filepath"
	"testing"
)

func TestNotificationImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "notification")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "notification")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{
		"SettingsReader", "EmailSender", "Enqueuer", "DashboardAlertReader", "TelegramSender", "LogRepository", "LogListFilter",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"EnqueueInput", "Service"})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "send.go"), []string{"TestSendInput"})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "log_service.go"), []string{"LogRecordInput", "LogService"})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "order_format.go"), []string{"OrderItemCounts"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "log_store.go"), []string{"LogStore"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"SettingsService", "LogService", "Sender", "AdminHandler",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})

	assertDirectoryGoFileBudget(t, moduleRoot, 13)
	assertDirectoryGoFileBudget(t, storeRoot, 3)
	assertDirectoryGoFileBudget(t, transportRoot, 3)
}

func TestNotificationLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	patterns := []string{
		filepath.Join(repositoryRoot, "internal", "service", "notification_service*.go"),
		filepath.Join(repositoryRoot, "internal", "service", "notification_log*.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "notification_log*.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_notification*.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		if len(matches) != 0 {
			t.Errorf("legacy notification files must stay removed: %v", matches)
		}
	}
}
