package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTelegramBroadcastHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "telegram")
	broadcastRoot := filepath.Join(moduleRoot, "broadcast")
	broadcastTransportRoot := filepath.Join(broadcastRoot, "transport", "http")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "telegram")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "notify.go"), []string{
		"NotifyService", "SettingReader", "SendOptions",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "notify.go"), []string{"NewNotifyService"})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterChannelBotRoutes",
	})
	assertFileDeclaresFunctions(t, filepath.Join(broadcastTransportRoot, "routes.go"), []string{
		"RegisterAdminRoutes",
	})
	assertFileDeclaresTypes(t, filepath.Join(broadcastTransportRoot, "handler.go"), []string{
		"AdminHandler", "AdminService",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "channel_bot_handler.go"), []string{
		"ChannelBotHandler", "BotSettings", "ChannelBotTokenProvider",
	})
	assertDirectoryGoFileBudget(t, moduleRoot, 3)
	assertDirectoryGoFileBudget(t, broadcastTransportRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 3)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_telegram_broadcast.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "channel", "channel_telegram_bot.go"),
		filepath.Join(repositoryRoot, "internal", "service", "telegram_notify_service.go"),
		filepath.Join(repositoryRoot, "internal", "models", "telegram_broadcast.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "telegram_broadcast_repository.go"),
		filepath.Join(repositoryRoot, "internal", "service", "telegram_broadcast_service.go"),
		filepath.Join(repositoryRoot, "internal", "transport", "http", "telegram", "admin_broadcast_handler.go"),
		filepath.Join(repositoryRoot, "internal", "wiring", "telegram", "broadcast.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy telegram handler must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy telegram handler: %v", err)
		}
	}
}
