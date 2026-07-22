package architecture

import (
	"path/filepath"
	"testing"
)

func TestMemberLevelImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "memberlevel")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "memberlevel")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "ports.go"), []string{
		"ListFilter", "LevelRepository", "PriceRepository", "UserRepository",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "service.go"), []string{
		"NewService", "CreateLevel", "UpdateLevel", "DeleteLevel", "ResolveMemberPrice",
		"CheckAndUpgrade", "OnRechargeCompleted", "OnOrderPaid", "AssignDefaultLevel",
		"SetUserLevel", "BackfillDefaultLevel",
	})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "level_store.go"), []string{"LevelStore"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "price_store.go"), []string{"PriceStore"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "user_store.go"), []string{"UserStore"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterAdminRoutes", "RegisterPublicRoutes", "RegisterChannelRoutes",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "list_handler.go"), []string{
		"NewPublicHandler", "NewChannelHandler",
	})

	assertDirectoryGoFileBudget(t, moduleRoot, 5)
	assertDirectoryGoFileBudget(t, storeRoot, 4)
	assertDirectoryGoFileBudget(t, transportRoot, 5)
}

func TestMemberLevelLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	patterns := []string{
		filepath.Join(repositoryRoot, "internal", "repository", "member_level*.go"),
		filepath.Join(repositoryRoot, "internal", "service", "member_level_service*.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_member_level.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		if len(matches) != 0 {
			t.Errorf("legacy member-level files must stay removed: %v", matches)
		}
	}
}
