package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTelegramAuthLivesInIdentityModule(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	identityRoot := filepath.Join(repositoryRoot, "internal", "modules", "identity")
	moduleRoot := filepath.Join(identityRoot, "telegramauth")
	applicationRoot := filepath.Join(moduleRoot, "application")

	assertFileDeclaresTypes(t, filepath.Join(applicationRoot, "service.go"), []string{
		"Service", "LoginPayload", "IdentityVerified", "Option",
	})
	assertFileDeclaresFunctions(t, filepath.Join(applicationRoot, "service.go"), []string{
		"NewService", "WithReplaySetNX", "WithOIDCStateStore",
	})
	assertDirectoryGoFileBudget(t, identityRoot, 0)
	assertDirectoryGoFileBudget(t, moduleRoot, 0)
	assertDirectoryGoFileBudget(t, applicationRoot, 5)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "service", "telegram_auth_service.go"),
		filepath.Join(repositoryRoot, "internal", "service", "telegram_auth_service_test.go"),
		filepath.Join(repositoryRoot, "internal", "service", "telegram_oidc.go"),
		filepath.Join(repositoryRoot, "internal", "service", "telegram_oidc_test.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy Telegram auth path must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy Telegram auth path: %v", err)
		}
	}
}

func TestUserProfileHTTPLivesInTransport(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "userauth")

	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterUserProfileRoutes",
		"RegisterUserEmailRoutes",
		"RegisterUserPasswordAuthRoutes",
		"RegisterUserPasswordRoutes",
		"RegisterUserVerifyAuthRoutes",
		"RegisterUserRegisterAuthRoutes",
		"RegisterUserLoginAuthRoutes",
		"RegisterUser2FAAuthRoutes",
		"RegisterUser2FARoutes",
		"RegisterUserTelegramOIDCAuthRoutes",
		"RegisterUserTelegramOIDCRoutes",
		"RegisterUserTelegramAuthRoutes",
		"RegisterUserTelegramRoutes",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_profile_handler.go"), []string{
		"UserProfileService", "UserProfileHandler", "UserProfileUpdateRequest",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_profile_handler.go"), []string{
		"NewUserProfileHandler", "GetCurrentUser", "userProfileResponse", "UpdateUserProfile",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_email_handler.go"), []string{
		"UserEmailService", "UserEmailHandler", "ChangeEmailSendCodeRequest", "ChangeEmailRequest",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_email_handler.go"), []string{
		"NewUserEmailHandler", "SendChangeEmailCode", "ChangeEmail", "changeEmailProfileResponse",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_password_handler.go"), []string{
		"UserPasswordService", "UserPasswordHandler", "UserResetPasswordRequest", "ChangeUserPasswordRequest", "WeakPasswordError",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_password_handler.go"), []string{
		"NewUserPasswordHandler", "UserForgotPassword", "ChangeUserPassword", "NewWeakPasswordError", "respondWeakPassword",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_verify_handler.go"), []string{
		"UserVerifySettings", "UserVerifyAuth", "UserVerifyHandler", "UserSendVerifyCodeRequest", "CaptchaVerifier",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_verify_handler.go"), []string{
		"NewUserVerifyHandler", "SendUserVerifyCode", "respondCaptchaError",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_telegram_oidc_handler.go"), []string{
		"UserTelegramOIDCService", "UserTelegramOIDCHandler", "AuthLoginResult", "LoginRecorder",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_telegram_oidc_handler.go"), []string{
		"NewUserTelegramOIDCHandler", "StartTelegramOIDCLogin", "TelegramOIDCLoginCallback",
		"StartTelegramOIDCBind", "TelegramOIDCBindCallback", "respondTelegramOIDCError",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_telegram_handler.go"), []string{
		"UserTelegramService", "UserTelegramHandler", "TelegramAuthPayload",
		"UserTelegramLoginRequest", "UserTelegramMiniAppAuthRequest", "UserBindTelegramRequest",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_telegram_handler.go"), []string{
		"NewUserTelegramHandler", "UserTelegramLogin", "UserTelegramMiniAppLogin",
		"GetMyTelegramBinding", "BindMyTelegram", "BindMyTelegramMiniApp", "UnbindMyTelegram",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_login_handler.go"), []string{
		"UserLoginSettings", "UserLoginAuth", "UserLoginHandler", "UserRegisterRequest", "UserLoginRequest",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_login_handler.go"), []string{
		"NewUserLoginHandler", "UserRegister", "UserLogin",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "user_2fa_handler.go"), []string{
		"User2FATOTPService", "User2FAAuthService", "User2FAChallengeStore", "User2FAHandler",
		"UserTOTPStatus", "UserTOTPSetupResult", "UserTOTPEnableResult", "UserChallengeClaims",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "user_2fa_handler.go"), []string{
		"NewUser2FAHandler", "GetUser2FAStatus", "SetupUser2FA", "EnableUser2FA",
		"DisableUser2FA", "RegenerateUser2FARecoveryCodes", "VerifyUser2FA",
	})
	assertDirectoryGoFileBudget(t, transportRoot, 10)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_profile.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_email.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_auth_password.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_auth_verify.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_auth_telegram_oidc.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_auth_telegram.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_auth_login.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_2fa.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "user_auth_login_record.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy userauth handler must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy userauth handler: %v", err)
		}
	}

	for _, legacy := range []string{"userauth_adapter.go", "giftcard_captcha_adapter.go"} {
		if _, err := os.Stat(filepath.Join(repositoryRoot, "internal", "router", legacy)); err == nil {
			t.Fatalf("%s belongs in internal/wiring, not internal/router", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy router adapter %s: %v", legacy, err)
		}
	}
	userAuthWiringRoot := filepath.Join(repositoryRoot, "internal", "wiring", "userauth")
	for _, file := range []string{"wiring.go", "adapters.go"} {
		if _, err := os.Stat(filepath.Join(userAuthWiringRoot, file)); err != nil {
			t.Fatalf("userauth wiring file %s missing: %v", file, err)
		}
	}
	assertDirectoryGoFileBudget(t, userAuthWiringRoot, 4)
	captchaWiringRoot := filepath.Join(repositoryRoot, "internal", "wiring", "captcha")
	if _, err := os.Stat(captchaWiringRoot); err == nil {
		t.Fatalf("captcha wiring must stay moved into the captcha module/bootstrap boundary: %s", captchaWiringRoot)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat retired captcha wiring: %v", err)
	}
}
