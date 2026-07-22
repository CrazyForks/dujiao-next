package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserAuthTelegramServiceIsSplitByFlow(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	legacyPath := filepath.Join(serviceDirectory, "user_auth_service_oauth.go")
	if _, err := os.Stat(legacyPath); err == nil {
		t.Fatalf("user_auth_service_oauth.go must be replaced by Telegram-flow-focused service files")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat user_auth_service_oauth.go: %v", err)
	}

	expectedOwner := map[string]string{
		"LoginWithTelegram":                     "user_auth_telegram_login.go",
		"LoginWithTelegramMiniApp":              "user_auth_telegram_login.go",
		"loginWithVerifiedTelegram":             "user_auth_telegram_login.go",
		"StartTelegramOIDC":                     "user_auth_telegram_oidc.go",
		"LoginWithTelegramOIDC":                 "user_auth_telegram_oidc.go",
		"BindTelegramOIDC":                      "user_auth_telegram_oidc.go",
		"BindTelegram":                          "user_auth_telegram_binding.go",
		"BindTelegramMiniApp":                   "user_auth_telegram_binding.go",
		"bindVerifiedTelegram":                  "user_auth_telegram_binding.go",
		"UnbindTelegram":                        "user_auth_telegram_binding.go",
		"GetTelegramBinding":                    "user_auth_telegram_binding.go",
		"ResolveTelegramChannelIdentity":        "user_auth_telegram_channel.go",
		"ProvisionTelegramChannelIdentity":      "user_auth_telegram_channel.go",
		"BindTelegramChannelByEmailCode":        "user_auth_telegram_channel.go",
		"resolveTelegramChannelIdentity":        "user_auth_telegram_channel.go",
		"provisionTelegramChannelIdentity":      "user_auth_telegram_channel.go",
		"bindTelegramIdentityToUser":            "user_auth_telegram_channel.go",
		"normalizeTelegramChannelIdentityInput": "user_auth_telegram_channel.go",
		"getActiveUserByID":                     "user_auth_telegram_identity.go",
		"findOrCreateTelegramUser":              "user_auth_telegram_identity.go",
		"getTelegramIdentityByVerifiedID":       "user_auth_telegram_identity.go",
		"canonicalizeTelegramProviderUserID":    "user_auth_telegram_identity.go",
		"telegramProviderUserIDMatchesVerified": "user_auth_telegram_identity.go",
		"applyTelegramIdentity":                 "user_auth_telegram_identity.go",
	}
	expectedTypeOwner := map[string]string{
		"LoginWithTelegramInput":              "user_auth_telegram_login.go",
		"LoginWithTelegramMiniAppInput":       "user_auth_telegram_login.go",
		"StartTelegramOIDCInput":              "user_auth_telegram_oidc.go",
		"LoginWithTelegramOIDCInput":          "user_auth_telegram_oidc.go",
		"BindTelegramOIDCInput":               "user_auth_telegram_oidc.go",
		"BindTelegramInput":                   "user_auth_telegram_binding.go",
		"BindTelegramMiniAppInput":            "user_auth_telegram_binding.go",
		"TelegramChannelIdentityInput":        "user_auth_telegram_channel.go",
		"BindTelegramChannelByEmailCodeInput": "user_auth_telegram_channel.go",
	}

	files := []string{
		"user_auth_telegram_login.go",
		"user_auth_telegram_oidc.go",
		"user_auth_telegram_binding.go",
		"user_auth_telegram_channel.go",
		"user_auth_telegram_identity.go",
	}
	actualOwners := make(map[string][]string, len(expectedOwner))
	actualTypeOwners := make(map[string][]string, len(expectedTypeOwner))
	for _, file := range files {
		parsed := parseProductionGoFile(t, filepath.Join(serviceDirectory, file))
		for _, function := range declaredFunctionNames(parsed) {
			if _, tracked := expectedOwner[function]; tracked {
				actualOwners[function] = append(actualOwners[function], file)
			}
		}
		for _, typeName := range declaredTypeNames(parsed) {
			if _, tracked := expectedTypeOwner[typeName]; tracked {
				actualTypeOwners[typeName] = append(actualTypeOwners[typeName], file)
			}
		}
	}

	for function, wantFile := range expectedOwner {
		gotFiles := actualOwners[function]
		if len(gotFiles) != 1 || gotFiles[0] != wantFile {
			t.Errorf("%s ownership mismatch: want [%s], got %v", function, wantFile, gotFiles)
		}
	}
	for typeName, wantFile := range expectedTypeOwner {
		gotFiles := actualTypeOwners[typeName]
		if len(gotFiles) != 1 || gotFiles[0] != wantFile {
			t.Errorf("%s ownership mismatch: want [%s], got %v", typeName, wantFile, gotFiles)
		}
	}
}
