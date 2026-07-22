package userauthwiring

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/auditlog"
	telegramauthapp "github.com/dujiao-next/internal/modules/identity/telegramauth/application"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	userauthtransport "github.com/dujiao-next/internal/transport/http/userauth"
)

// userProfileTransportAdapter 将用户认证服务适配为用户资料 transport 端口。
type userProfileTransportAdapter struct {
	service *service.UserAuthService
}

func (a userProfileTransportAdapter) GetUserByID(id uint) (*models.User, error) {
	user, err := a.service.GetUserByID(id)
	return user, mapUserAuthTransportError(err)
}

func (a userProfileTransportAdapter) ResolveEmailChangeMode(user *models.User) (string, error) {
	mode, err := a.service.ResolveEmailChangeMode(user)
	return mode, mapUserAuthTransportError(err)
}

func (a userProfileTransportAdapter) ResolvePasswordChangeMode(user *models.User) (string, error) {
	mode, err := a.service.ResolvePasswordChangeMode(user)
	return mode, mapUserAuthTransportError(err)
}

func (a userProfileTransportAdapter) UpdateProfile(userID uint, nickname, locale *string) (*models.User, error) {
	user, err := a.service.UpdateProfile(userID, nickname, locale)
	return user, mapUserAuthTransportError(err)
}

// userEmailTransportAdapter 将用户认证服务适配为更换邮箱 transport 端口。
type userEmailTransportAdapter struct {
	service *service.UserAuthService
}

func (a userEmailTransportAdapter) SendChangeEmailCode(userID uint, kind, newEmail, locale string) error {
	return mapUserAuthTransportError(a.service.SendChangeEmailCode(userID, kind, newEmail, locale))
}

func (a userEmailTransportAdapter) ChangeEmail(userID uint, newEmail, oldCode, newCode string) (*models.User, error) {
	user, err := a.service.ChangeEmail(userID, newEmail, oldCode, newCode)
	return user, mapUserAuthTransportError(err)
}

func (a userEmailTransportAdapter) ResolveEmailChangeMode(user *models.User) (string, error) {
	mode, err := a.service.ResolveEmailChangeMode(user)
	return mode, mapUserAuthTransportError(err)
}

func (a userEmailTransportAdapter) ResolvePasswordChangeMode(user *models.User) (string, error) {
	mode, err := a.service.ResolvePasswordChangeMode(user)
	return mode, mapUserAuthTransportError(err)
}

// userPasswordTransportAdapter 将用户认证/设置服务适配为密码 transport 端口。
type userPasswordTransportAdapter struct {
	auth     *service.UserAuthService
	settings *settingsapp.Service
}

func (a userPasswordTransportAdapter) GetEmailVerificationEnabled(defaultValue bool) (bool, error) {
	return a.settings.GetEmailVerificationEnabled(defaultValue)
}

func (a userPasswordTransportAdapter) ResetPassword(email, code, newPassword string) error {
	return mapUserAuthTransportError(a.auth.ResetPassword(email, code, newPassword))
}

func (a userPasswordTransportAdapter) ChangePassword(userID uint, oldPassword, newPassword string) error {
	return mapUserAuthTransportError(a.auth.ChangePassword(userID, oldPassword, newPassword))
}

// userVerifyTransportAdapter 将用户认证/设置服务适配为验证码发送端口。
type userVerifyTransportAdapter struct {
	auth     *service.UserAuthService
	settings *settingsapp.Service
}

func (a userVerifyTransportAdapter) GetEmailVerificationEnabled(defaultValue bool) (bool, error) {
	return a.settings.GetEmailVerificationEnabled(defaultValue)
}

func (a userVerifyTransportAdapter) GetRegistrationEnabled(defaultValue bool) (bool, error) {
	return a.settings.GetRegistrationEnabled(defaultValue)
}

func (a userVerifyTransportAdapter) SendVerifyCode(email, purpose, locale string) error {
	return mapUserAuthTransportError(a.auth.SendVerifyCode(email, purpose, locale))
}

// userTelegramTransportAdapter 将用户认证服务适配为 Telegram widget/MiniApp transport 端口。
type userTelegramTransportAdapter struct {
	auth *service.UserAuthService
}

func (a userTelegramTransportAdapter) toServicePayload(payload userauthtransport.TelegramAuthPayload) telegramauthapp.LoginPayload {
	return telegramauthapp.LoginPayload{
		ID:        payload.ID,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Username:  payload.Username,
		PhotoURL:  payload.PhotoURL,
		AuthDate:  payload.AuthDate,
		Hash:      payload.Hash,
	}
}

func (a userTelegramTransportAdapter) toAuthLoginResult(res *service.UserLoginResult) *userauthtransport.AuthLoginResult {
	if res == nil {
		return nil
	}
	return &userauthtransport.AuthLoginResult{
		RequiresTOTP:       res.RequiresTOTP,
		User:               res.User,
		Token:              res.Token,
		ExpiresAt:          res.ExpiresAt,
		ChallengeToken:     res.ChallengeToken,
		ChallengeExpiresAt: res.ChallengeExpiresAt,
	}
}

func (a userTelegramTransportAdapter) LoginWithTelegram(ctx context.Context, payload userauthtransport.TelegramAuthPayload) (*userauthtransport.AuthLoginResult, error) {
	res, err := a.auth.LoginWithTelegram(service.LoginWithTelegramInput{
		Payload: a.toServicePayload(payload),
		Context: ctx,
	})
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	return a.toAuthLoginResult(res), nil
}

func (a userTelegramTransportAdapter) LoginWithTelegramMiniApp(ctx context.Context, initData string) (*userauthtransport.AuthLoginResult, error) {
	res, err := a.auth.LoginWithTelegramMiniApp(service.LoginWithTelegramMiniAppInput{
		InitData: initData,
		Context:  ctx,
	})
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	return a.toAuthLoginResult(res), nil
}

func (a userTelegramTransportAdapter) GetTelegramBinding(userID uint) (*models.UserOAuthIdentity, error) {
	identity, err := a.auth.GetTelegramBinding(userID)
	return identity, mapUserAuthTransportError(err)
}

func (a userTelegramTransportAdapter) BindTelegram(ctx context.Context, userID uint, payload userauthtransport.TelegramAuthPayload) (*models.UserOAuthIdentity, error) {
	identity, err := a.auth.BindTelegram(service.BindTelegramInput{
		UserID:  userID,
		Payload: a.toServicePayload(payload),
		Context: ctx,
	})
	return identity, mapUserAuthTransportError(err)
}

func (a userTelegramTransportAdapter) BindTelegramMiniApp(ctx context.Context, userID uint, initData string) (*models.UserOAuthIdentity, error) {
	identity, err := a.auth.BindTelegramMiniApp(service.BindTelegramMiniAppInput{
		UserID:   userID,
		InitData: initData,
		Context:  ctx,
	})
	return identity, mapUserAuthTransportError(err)
}

func (a userTelegramTransportAdapter) UnbindTelegram(userID uint) error {
	return mapUserAuthTransportError(a.auth.UnbindTelegram(userID))
}

// userTelegramOIDCTransportAdapter 将用户认证服务适配为 Telegram OIDC transport 端口。
type userTelegramOIDCTransportAdapter struct {
	auth *service.UserAuthService
}

func (a userTelegramOIDCTransportAdapter) StartTelegramOIDC(ctx context.Context, intent string, userID uint) (string, error) {
	authURL, err := a.auth.StartTelegramOIDC(service.StartTelegramOIDCInput{
		Intent:  intent,
		UserID:  userID,
		Context: ctx,
	})
	return authURL, mapUserAuthTransportError(err)
}

func (a userTelegramOIDCTransportAdapter) LoginWithTelegramOIDC(ctx context.Context, code, state string) (*userauthtransport.AuthLoginResult, error) {
	res, err := a.auth.LoginWithTelegramOIDC(service.LoginWithTelegramOIDCInput{
		Code:    code,
		State:   state,
		Context: ctx,
	})
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	if res == nil {
		return nil, nil
	}
	return &userauthtransport.AuthLoginResult{
		RequiresTOTP:       res.RequiresTOTP,
		User:               res.User,
		Token:              res.Token,
		ExpiresAt:          res.ExpiresAt,
		ChallengeToken:     res.ChallengeToken,
		ChallengeExpiresAt: res.ChallengeExpiresAt,
	}, nil
}

func (a userTelegramOIDCTransportAdapter) BindTelegramOIDC(ctx context.Context, userID uint, code, state string) (*models.UserOAuthIdentity, error) {
	identity, err := a.auth.BindTelegramOIDC(service.BindTelegramOIDCInput{
		UserID:  userID,
		Code:    code,
		State:   state,
		Context: ctx,
	})
	return identity, mapUserAuthTransportError(err)
}

// userLoginTransportAdapter 将设置/认证服务适配为注册登录 transport 端口。
type userLoginTransportAdapter struct {
	auth     *service.UserAuthService
	settings *settingsapp.Service
}

func (a userLoginTransportAdapter) GetRegistrationEnabled(defaultValue bool) (bool, error) {
	return a.settings.GetRegistrationEnabled(defaultValue)
}

func (a userLoginTransportAdapter) GetEmailVerificationEnabled(defaultValue bool) (bool, error) {
	return a.settings.GetEmailVerificationEnabled(defaultValue)
}

func (a userLoginTransportAdapter) Register(email, password, code string, agreementAccepted, emailVerificationEnabled bool) (*models.User, string, time.Time, error) {
	user, token, expiresAt, err := a.auth.Register(email, password, code, agreementAccepted, emailVerificationEnabled)
	return user, token, expiresAt, mapUserAuthTransportError(err)
}

func (a userLoginTransportAdapter) LoginStep1(email, password string, rememberMe bool) (*userauthtransport.AuthLoginResult, error) {
	res, err := a.auth.LoginStep1(email, password, rememberMe)
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	if res == nil {
		return nil, nil
	}
	return &userauthtransport.AuthLoginResult{
		RequiresTOTP:       res.RequiresTOTP,
		User:               res.User,
		Token:              res.Token,
		ExpiresAt:          res.ExpiresAt,
		ChallengeToken:     res.ChallengeToken,
		ChallengeExpiresAt: res.ChallengeExpiresAt,
	}, nil
}

// userLoginRecorderAdapter 将登录日志服务适配为 transport 记录端口。
type userLoginRecorderAdapter struct {
	logs *auditlog.UserLoginService
}

func (a userLoginRecorderAdapter) Record(email string, userID uint, status, failReason, source, clientIP, userAgent, requestID string) {
	if a.logs == nil {
		return
	}
	_ = a.logs.Record(auditlog.UserLoginRecord{
		UserID:      userID,
		Email:       email,
		Status:      status,
		FailReason:  failReason,
		ClientIP:    clientIP,
		UserAgent:   userAgent,
		LoginSource: source,
		RequestID:   strings.TrimSpace(requestID),
	})
}

// user2FATOTPTransportAdapter 将用户 TOTP 服务适配为 2FA transport 端口。
type user2FATOTPTransportAdapter struct {
	totp *service.UserTOTPService
}

func (a user2FATOTPTransportAdapter) GetStatus(userID uint) (*userauthtransport.UserTOTPStatus, error) {
	st, err := a.totp.GetStatus(userID)
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	if st == nil {
		return nil, nil
	}
	return &userauthtransport.UserTOTPStatus{
		Enabled:                st.Enabled,
		EnabledAt:              st.EnabledAt,
		RecoveryCodesRemaining: st.RecoveryCodesRemaining,
		RecoveryCodesTotal:     st.RecoveryCodesTotal,
	}, nil
}

func (a user2FATOTPTransportAdapter) Setup(userID uint) (*userauthtransport.UserTOTPSetupResult, error) {
	res, err := a.totp.Setup(userID)
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	if res == nil {
		return nil, nil
	}
	return &userauthtransport.UserTOTPSetupResult{
		Secret:     res.Secret,
		OtpauthURL: res.OtpauthURL,
		ExpiresAt:  res.ExpiresAt,
	}, nil
}

func (a user2FATOTPTransportAdapter) Enable(userID uint, code string) (*userauthtransport.UserTOTPEnableResult, error) {
	res, err := a.totp.Enable(userID, code)
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	if res == nil {
		return nil, nil
	}
	return &userauthtransport.UserTOTPEnableResult{
		EnabledAt:     res.EnabledAt,
		RecoveryCodes: res.RecoveryCodes,
		Token:         res.Token,
		ExpiresAt:     res.ExpiresAt,
	}, nil
}

func (a user2FATOTPTransportAdapter) Disable(userID uint, code string, isRecoveryCode bool) error {
	return mapUserAuthTransportError(a.totp.Disable(userID, code, isRecoveryCode))
}

func (a user2FATOTPTransportAdapter) RegenerateRecoveryCodes(userID uint, code string) ([]string, error) {
	codes, err := a.totp.RegenerateRecoveryCodes(userID, code)
	return codes, mapUserAuthTransportError(err)
}

func (a user2FATOTPTransportAdapter) VerifyChallengeCode(userID uint, code string) error {
	return mapUserAuthTransportError(a.totp.VerifyChallengeCode(userID, code))
}

func (a user2FATOTPTransportAdapter) VerifyChallengeRecoveryCode(userID uint, code string) error {
	return mapUserAuthTransportError(a.totp.VerifyChallengeRecoveryCode(userID, code))
}

// user2FAAuthTransportAdapter 将用户认证/仓储适配为 2FA 登录完成端口。
type user2FAAuthTransportAdapter struct {
	auth  *service.UserAuthService
	users repository.UserRepository
}

func (a user2FAAuthTransportAdapter) ParseUserChallengeToken(tokenString string) (*userauthtransport.UserChallengeClaims, error) {
	claims, err := a.auth.ParseUserChallengeToken(tokenString)
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	if claims == nil {
		return nil, nil
	}
	return &userauthtransport.UserChallengeClaims{
		UserID:     claims.UserID,
		JTI:        claims.JTI,
		RememberMe: claims.RememberMe,
	}, nil
}

func (a user2FAAuthTransportAdapter) CompleteLoginAfter2FA(userID uint, rememberMe bool) (*userauthtransport.AuthLoginResult, error) {
	res, err := a.auth.CompleteLoginAfter2FA(userID, rememberMe)
	if err != nil {
		return nil, mapUserAuthTransportError(err)
	}
	if res == nil {
		return nil, nil
	}
	return &userauthtransport.AuthLoginResult{
		RequiresTOTP:       res.RequiresTOTP,
		User:               res.User,
		Token:              res.Token,
		ExpiresAt:          res.ExpiresAt,
		ChallengeToken:     res.ChallengeToken,
		ChallengeExpiresAt: res.ChallengeExpiresAt,
	}, nil
}

func (a user2FAAuthTransportAdapter) GetUserEmail(userID uint) (string, error) {
	user, err := a.users.GetByID(userID)
	if err != nil {
		return "", mapUserAuthTransportError(err)
	}
	if user == nil {
		return "", nil
	}
	return user.Email, nil
}

// user2FAChallengeStoreAdapter 将 Redis 挑战状态适配为 transport 端口。
type user2FAChallengeStoreAdapter struct{}

func (user2FAChallengeStoreAdapter) IsRevoked(ctx context.Context, jti string) bool {
	rdb := cache.Client()
	if rdb == nil {
		return false
	}
	v, _ := rdb.Exists(ctx, service.UserChallengeRevokedKey(jti)).Result()
	return v == 1
}

func (user2FAChallengeStoreAdapter) BumpFails(ctx context.Context, jti string) int64 {
	rdb := cache.Client()
	if rdb == nil {
		return 0
	}
	cnt, err := rdb.Incr(ctx, service.UserChallengeFailKey(jti)).Result()
	if err == nil && cnt == 1 {
		_ = rdb.Expire(ctx, service.UserChallengeFailKey(jti), service.UserChallengeTTL).Err()
	}
	return cnt
}

func (user2FAChallengeStoreAdapter) Revoke(ctx context.Context, jti string) {
	rdb := cache.Client()
	if rdb == nil {
		return
	}
	_ = rdb.Set(ctx, service.UserChallengeRevokedKey(jti), "1", service.UserChallengeTTL).Err()
}

func mapUserAuthTransportError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, service.ErrWeakPassword) {
		type keyed interface {
			Key() string
			Args() []interface{}
		}
		var k keyed
		if errors.As(err, &k) {
			return userauthtransport.NewWeakPasswordError(k.Key(), k.Args()...)
		}
		if perr, ok := err.(keyed); ok {
			return userauthtransport.NewWeakPasswordError(perr.Key(), perr.Args()...)
		}
		return fmt.Errorf("%w: %v", userauthtransport.ErrWeakPassword, err)
	}
	for _, mapping := range []struct {
		source error
		target error
	}{
		{service.ErrProfileEmpty, userauthtransport.ErrProfileEmpty},
		{service.ErrNotFound, userauthtransport.ErrUserNotFound},
		{service.ErrInvalidEmail, userauthtransport.ErrInvalidEmail},
		{service.ErrEmailChangeInvalid, userauthtransport.ErrEmailChangeInvalid},
		{service.ErrEmailChangeExists, userauthtransport.ErrEmailChangeExists},
		{service.ErrVerifyCodeInvalid, userauthtransport.ErrVerifyCodeInvalid},
		{service.ErrVerifyCodeExpired, userauthtransport.ErrVerifyCodeExpired},
		{service.ErrVerifyCodeTooFrequent, userauthtransport.ErrVerifyCodeTooFrequent},
		{service.ErrVerifyCodeAttemptsExceeded, userauthtransport.ErrVerifyCodeAttemptsExceeded},
		{service.ErrEmailServiceDisabled, userauthtransport.ErrEmailServiceDisabled},
		{service.ErrEmailServiceNotConfigured, userauthtransport.ErrEmailServiceNotConfigured},
		{service.ErrEmailRecipientRejected, userauthtransport.ErrEmailRecipientRejected},
		{service.ErrInvalidPassword, userauthtransport.ErrInvalidPassword},
		{service.ErrInvalidVerifyPurpose, userauthtransport.ErrInvalidVerifyPurpose},
		{service.ErrEmailExists, userauthtransport.ErrEmailExists},
		{settingsapp.ErrEmailDomainNotAllowed, userauthtransport.ErrEmailDomainNotAllowed},
		{telegramauthapp.ErrTelegramAuthDisabled, userauthtransport.ErrTelegramAuthDisabled},
		{telegramauthapp.ErrTelegramAuthConfigInvalid, userauthtransport.ErrTelegramAuthConfigInvalid},
		{telegramauthapp.ErrTelegramOIDCStateInvalid, userauthtransport.ErrTelegramOIDCStateInvalid},
		{telegramauthapp.ErrTelegramOIDCTokenExchange, userauthtransport.ErrTelegramOIDCTokenExchange},
		{telegramauthapp.ErrTelegramOIDCIDTokenInvalid, userauthtransport.ErrTelegramOIDCIDTokenInvalid},
		{telegramauthapp.ErrTelegramAuthPayloadInvalid, userauthtransport.ErrTelegramAuthPayloadInvalid},
		{telegramauthapp.ErrTelegramAuthSignatureInvalid, userauthtransport.ErrTelegramAuthSignatureInvalid},
		{telegramauthapp.ErrTelegramAuthExpired, userauthtransport.ErrTelegramAuthExpired},
		{telegramauthapp.ErrTelegramAuthReplay, userauthtransport.ErrTelegramAuthReplay},
		{service.ErrUserOAuthIdentityExists, userauthtransport.ErrUserOAuthIdentityExists},
		{service.ErrUserOAuthAlreadyBound, userauthtransport.ErrUserOAuthAlreadyBound},
		{service.ErrUserOAuthNotBound, userauthtransport.ErrUserOAuthNotBound},
		{service.ErrTelegramUnbindRequiresEmail, userauthtransport.ErrTelegramUnbindRequiresEmail},
		{service.ErrUserDisabled, userauthtransport.ErrUserDisabled},
		{service.ErrRegistrationDisabled, userauthtransport.ErrRegistrationDisabled},
		{service.ErrAgreementRequired, userauthtransport.ErrAgreementRequired},
		{service.ErrInvalidCredentials, userauthtransport.ErrInvalidCredentials},
		{service.ErrEmailNotVerified, userauthtransport.ErrEmailNotVerified},
		{service.ErrTOTPAlreadyEnabled, userauthtransport.ErrTOTPAlreadyEnabled},
		{service.ErrTOTPNotEnabled, userauthtransport.ErrTOTPNotEnabled},
		{service.ErrTOTPPendingExpired, userauthtransport.ErrTOTPPendingExpired},
		{service.ErrTOTPCodeInvalid, userauthtransport.ErrTOTPCodeInvalid},
		{service.ErrTOTPRecoveryInvalid, userauthtransport.ErrTOTPRecoveryInvalid},
		{service.ErrTOTPTooManyAttempts, userauthtransport.ErrTOTPTooManyAttempts},
	} {
		if errors.Is(err, mapping.source) {
			return fmt.Errorf("%w: %v", mapping.target, err)
		}
	}
	return err
}
