package architecture

import (
	"path/filepath"
	"testing"
)

func TestAuditLogImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "auditlog")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "auditlog")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{
		"UserLoginFilter", "AuthzFilter", "AdminLoginFilter",
		"UserLoginRepository", "AuthzRepository", "AdminLoginRepository",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "user_login_service.go"), []string{
		"UserLoginRecord", "UserLoginService",
	})
	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "authz_service.go"), []string{
		"AuthzRecord", "AuthzService",
	})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "user_login_store.go"), []string{"UserLoginStore"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "authz_store.go"), []string{"AuthzStore"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "admin_login_store.go"), []string{"AdminLoginStore"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterAdminRoutes", "RegisterUserRoutes",
	})

	assertDirectoryGoFileBudget(t, moduleRoot, 4)
	assertDirectoryGoFileBudget(t, storeRoot, 5)
	assertDirectoryGoFileBudget(t, transportRoot, 4)
}

func TestAuditLogLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	patterns := []string{
		filepath.Join(repositoryRoot, "internal", "service", "user_login_log_service.go"),
		filepath.Join(repositoryRoot, "internal", "service", "authz_audit_service.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "user_login_log_repository.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "authz_audit_log_repository.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "admin_login_log_repository*.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_authz_audit.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "user_login_log_admin.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_login_log.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		if len(matches) != 0 {
			t.Errorf("legacy audit-log files must stay removed: %v", matches)
		}
	}
}
