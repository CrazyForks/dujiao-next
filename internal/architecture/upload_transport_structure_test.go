package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadAdminHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "upload")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "upload")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{"Service", "Result", "UploadValidationError"})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterAdminRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_handler.go"), []string{
		"FileUploader", "MediaRecorder", "AdminHandler",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 4)

	for _, relativePath := range []string{
		"internal/http/handlers/admin/admin_upload.go",
		"internal/http/handlers/admin/admin_upload_test.go",
		"internal/service/upload_service.go",
		"internal/wiring/upload/wiring.go",
	} {
		path := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
		if _, err := os.Stat(path); err == nil {
			t.Errorf("legacy upload handler must stay removed: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect %s: %v", relativePath, err)
		}
	}
}
