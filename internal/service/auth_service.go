package service

import (
	"context"
	"errors"
	"time"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	cfg       *config.Config
	adminRepo repository.AdminRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService(cfg *config.Config, adminRepo repository.AdminRepository) *AuthService {
	return &AuthService{
		cfg:       cfg,
		adminRepo: adminRepo,
	}
}

// HashPassword 使用 bcrypt 加密密码
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 验证密码
func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePassword 校验密码是否符合策略
func (s *AuthService) ValidatePassword(password string) error {
	if s == nil || s.cfg == nil {
		return nil
	}
	return ValidatePasswordPolicy(s.cfg.Security.PasswordPolicy, password)
}

// JWT typ 常量
const (
	TokenTypAccess       = "access"
	TokenTyp2FAChallenge = "2fa_challenge"
)

// IsAccessTokenTyp 判断 typ 是否为合法访问 token（空字符串兼容旧 token）
func IsAccessTokenTyp(typ string) bool {
	return typ == "" || typ == TokenTypAccess
}

// JWTClaims JWT 声明
type JWTClaims struct {
	AdminID      uint   `json:"admin_id"`
	Username     string `json:"username"`
	TokenVersion uint64 `json:"token_version"`
	Typ          string `json:"typ,omitempty"`
	jwt.RegisteredClaims
}

// ChallengeClaims 2FA 挑战 token 的 JWT claims
//
// 注：Typ 字段同时占用与 JWTClaims 兼容的 typ 键，写入 "2fa_challenge"，
// 防止挑战 token 在被错误地解析为 JWTClaims 时通过中间件的 typ 校验。
type ChallengeClaims struct {
	AdminID uint   `json:"admin_id"`
	JTI     string `json:"jti"`
	Purpose string `json:"purpose"`
	Typ     string `json:"typ"`
	jwt.RegisteredClaims
}

// LoginResult 登录第一步结果
type LoginResult struct {
	RequiresTOTP       bool
	Admin              *models.Admin
	Token              string
	ExpiresAt          time.Time
	ChallengeToken     string
	ChallengeJTI       string
	ChallengeExpiresAt time.Time
}

// ChallengePurpose2FA 挑战 token 的 purpose 常量
const ChallengePurpose2FA = "2fa_challenge"

// ChallengeTokenTTL 挑战 token 有效期
const ChallengeTokenTTL = 5 * time.Minute

// GenerateJWT 生成 JWT Token
func (s *AuthService) GenerateJWT(admin *models.Admin) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.JWT.ExpireHours) * time.Hour)

	claims := JWTClaims{
		AdminID:      admin.ID,
		Username:     admin.Username,
		TokenVersion: admin.TokenVersion,
		Typ:          TokenTypAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.SecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ParseJWT 解析 JWT Token
func (s *AuthService) ParseJWT(tokenString string) (*JWTClaims, error) {
	parser := newHS256JWTParser()
	token, err := parser.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的 token")
}

// Login 管理员登录（第一步）
func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	admin, err := s.adminRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyhashtopreventtimingattacksxxxxxxxxxxxxxxxxxx"), []byte(password))
		return nil, ErrInvalidCredentials
	}
	if err := s.VerifyPassword(admin.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 已启用 2FA → 仅签发挑战 token
	if admin.TOTPEnabledAt != nil {
		challenge, jti, expiresAt, err := s.IssueChallengeToken(admin.ID)
		if err != nil {
			return nil, err
		}
		return &LoginResult{
			RequiresTOTP:       true,
			Admin:              admin,
			ChallengeToken:     challenge,
			ChallengeJTI:       jti,
			ChallengeExpiresAt: expiresAt,
		}, nil
	}

	// 未启用 → 直接发正式 JWT
	token, expiresAt, err := s.GenerateJWT(admin)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	admin.LastLoginAt = &now
	if err := s.adminRepo.Update(admin); err != nil {
		return nil, err
	}
	_ = cache.SetAdminAuthState(context.Background(), cache.BuildAdminAuthState(admin))
	return &LoginResult{
		RequiresTOTP: false,
		Admin:        admin,
		Token:        token,
		ExpiresAt:    expiresAt,
	}, nil
}

// IssueChallengeToken 签发 2FA 挑战 token
func (s *AuthService) IssueChallengeToken(adminID uint) (token, jti string, expiresAt time.Time, err error) {
	jti = uuid.NewString()
	expiresAt = time.Now().Add(ChallengeTokenTTL)
	claims := ChallengeClaims{
		AdminID: adminID,
		JTI:     jti,
		Purpose: ChallengePurpose2FA,
		Typ:     TokenTyp2FAChallenge,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(s.cfg.JWT.SecretKey))
	if err != nil {
		return "", "", time.Time{}, err
	}
	return signed, jti, expiresAt, nil
}

// ParseChallengeToken 解析并校验挑战 token
func (s *AuthService) ParseChallengeToken(tokenString string) (*ChallengeClaims, error) {
	parser := newHS256JWTParser()
	tok, err := parser.ParseWithClaims(tokenString, &ChallengeClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*ChallengeClaims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid challenge token")
	}
	if claims.Purpose != ChallengePurpose2FA || claims.Typ != TokenTyp2FAChallenge {
		return nil, errors.New("invalid challenge purpose")
	}
	return claims, nil
}

// CompleteLoginAfter2FA 在 2FA 验证通过后完成登录：发正式 JWT、更新 last_login
func (s *AuthService) CompleteLoginAfter2FA(adminID uint) (*LoginResult, error) {
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrNotFound
	}
	token, expiresAt, err := s.GenerateJWT(admin)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	admin.LastLoginAt = &now
	if err := s.adminRepo.Update(admin); err != nil {
		return nil, err
	}
	_ = cache.SetAdminAuthState(context.Background(), cache.BuildAdminAuthState(admin))
	return &LoginResult{RequiresTOTP: false, Admin: admin, Token: token, ExpiresAt: expiresAt}, nil
}

// AdminRepo 暴露给 handler 用（例如 2FA reset 后查 username）
func (s *AuthService) AdminRepo() repository.AdminRepository {
	return s.adminRepo
}

// ChangePassword 修改管理员密码
func (s *AuthService) ChangePassword(adminID uint, oldPassword, newPassword string) error {
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return ErrNotFound
	}

	if err := s.VerifyPassword(admin.PasswordHash, oldPassword); err != nil {
		return ErrInvalidPassword
	}

	if err := s.ValidatePassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	admin.PasswordHash = hashedPassword
	now := time.Now()
	admin.TokenVersion++
	admin.TokenInvalidBefore = &now
	if err := s.adminRepo.Update(admin); err != nil {
		return err
	}
	_ = cache.SetAdminAuthState(context.Background(), cache.BuildAdminAuthState(admin))
	return nil
}
