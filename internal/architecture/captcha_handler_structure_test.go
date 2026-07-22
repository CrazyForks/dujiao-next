package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPublicCaptchaHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "captcha")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "captcha")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "service.go"), []string{
		"Service", "SettingReader", "VerifyPayload", "ImageChallenge",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{"NewService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{"RegisterPublicRoutes"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "public_handler.go"), []string{
		"PublicHandler", "ImageChallengeGenerator",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 4)

	legacy := filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "captcha.go")
	if _, err := os.Stat(legacy); err == nil {
		t.Fatalf("legacy public captcha handler must stay removed: %s", legacy)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat legacy captcha handler: %v", err)
	}
	for _, relativePath := range []string{
		"internal/service/captcha_service.go",
		"internal/service/captcha_setting.go",
	} {
		path := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("legacy captcha service file must stay removed: %s", relativePath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", relativePath, err)
		}
	}
}
