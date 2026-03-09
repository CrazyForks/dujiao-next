package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/dujiao-next/internal/crypto"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/upstream"
)

// ChannelClientService 渠道客户端业务服务
type ChannelClientService struct {
	repo      repository.ChannelClientRepository
	encKey    []byte // AES-256 密钥
	secretKey string // 原始密钥（用于派生）
}

// NewChannelClientService 创建渠道客户端服务
func NewChannelClientService(repo repository.ChannelClientRepository, appSecretKey string) *ChannelClientService {
	return &ChannelClientService{
		repo:      repo,
		encKey:    crypto.DeriveKey(appSecretKey),
		secretKey: appSecretKey,
	}
}

// ChannelClientResponse 渠道客户端响应（含明文 secret）
type ChannelClientResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	ChannelType   string `json:"channel_type"`
	ChannelKey    string `json:"channel_key"`
	ChannelSecret string `json:"channel_secret"`
	BotToken      string `json:"bot_token"`
	BotTokenSet   bool   `json:"bot_token_set"`
	Description   string `json:"description"`
	Status        int    `json:"status"`
}

// CreateChannelClient 创建渠道客户端
func (s *ChannelClientService) CreateChannelClient(name, channelType, description, botToken string) (*ChannelClientResponse, error) {
	// 生成随机 key (32 bytes = 64 hex chars)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("generate channel key: %w", err)
	}
	channelKey := hex.EncodeToString(keyBytes)

	// 生成随机 secret (32 bytes = 64 hex chars)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("generate channel secret: %w", err)
	}
	plainSecret := hex.EncodeToString(secretBytes)

	// 加密 secret 存储
	encryptedSecret, err := crypto.Encrypt(s.encKey, plainSecret)
	if err != nil {
		return nil, fmt.Errorf("encrypt channel secret: %w", err)
	}

	client := &models.ChannelClient{
		Name:          name,
		ChannelType:   channelType,
		ChannelKey:    channelKey,
		ChannelSecret: encryptedSecret,
		Status:        1,
		Description:   description,
	}

	// 加密 bot_token（如果提供）
	if botToken != "" {
		encryptedToken, err := crypto.Encrypt(s.encKey, botToken)
		if err != nil {
			return nil, fmt.Errorf("encrypt bot token: %w", err)
		}
		client.BotToken = encryptedToken
	}

	if err := s.repo.Create(client); err != nil {
		return nil, err
	}

	return &ChannelClientResponse{
		ID:            client.ID,
		Name:          client.Name,
		ChannelType:   client.ChannelType,
		ChannelKey:    client.ChannelKey,
		ChannelSecret: plainSecret,
		BotToken:      maskBotToken(botToken),
		BotTokenSet:   botToken != "",
		Description:   client.Description,
		Status:        client.Status,
	}, nil
}

// GetChannelClient 获取渠道客户端
func (s *ChannelClientService) GetChannelClient(id uint) (*models.ChannelClient, error) {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrChannelClientNotFound
	}
	return client, nil
}

// ListChannelClients 列出所有渠道客户端
func (s *ChannelClientService) ListChannelClients() ([]models.ChannelClient, error) {
	return s.repo.FindAll()
}

// GetChannelClientDetail 获取渠道客户端详情（含解密 secret）
func (s *ChannelClientService) GetChannelClientDetail(id uint) (*ChannelClientResponse, error) {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrChannelClientNotFound
	}

	plainSecret, err := crypto.Decrypt(s.encKey, client.ChannelSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt channel secret: %w", err)
	}

	resp := &ChannelClientResponse{
		ID:            client.ID,
		Name:          client.Name,
		ChannelType:   client.ChannelType,
		ChannelKey:    client.ChannelKey,
		ChannelSecret: plainSecret,
		BotTokenSet:   client.BotToken != "",
		Description:   client.Description,
		Status:        client.Status,
	}

	if client.BotToken != "" {
		plainToken, err := crypto.Decrypt(s.encKey, client.BotToken)
		if err == nil {
			resp.BotToken = maskBotToken(plainToken)
		}
	}

	return resp, nil
}

// ListChannelClientDetails 列出所有渠道客户端（含解密 secret）
func (s *ChannelClientService) ListChannelClientDetails() ([]ChannelClientResponse, error) {
	clients, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]ChannelClientResponse, 0, len(clients))
	for _, c := range clients {
		plainSecret, decErr := crypto.Decrypt(s.encKey, c.ChannelSecret)
		if decErr != nil {
			plainSecret = ""
		}
		resp := ChannelClientResponse{
			ID:            c.ID,
			Name:          c.Name,
			ChannelType:   c.ChannelType,
			ChannelKey:    c.ChannelKey,
			ChannelSecret: plainSecret,
			BotTokenSet:   c.BotToken != "",
			Description:   c.Description,
			Status:        c.Status,
		}
		if c.BotToken != "" {
			plainToken, decErr := crypto.Decrypt(s.encKey, c.BotToken)
			if decErr == nil {
				resp.BotToken = maskBotToken(plainToken)
			}
		}
		result = append(result, resp)
	}
	return result, nil
}

// ResetChannelClientSecret 重置渠道客户端 Secret
func (s *ChannelClientService) ResetChannelClientSecret(id uint) (*ChannelClientResponse, error) {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrChannelClientNotFound
	}

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("generate channel secret: %w", err)
	}
	plainSecret := hex.EncodeToString(secretBytes)

	encryptedSecret, err := crypto.Encrypt(s.encKey, plainSecret)
	if err != nil {
		return nil, fmt.Errorf("encrypt channel secret: %w", err)
	}

	client.ChannelSecret = encryptedSecret
	if err := s.repo.Update(client); err != nil {
		return nil, err
	}

	resp := &ChannelClientResponse{
		ID:            client.ID,
		Name:          client.Name,
		ChannelType:   client.ChannelType,
		ChannelKey:    client.ChannelKey,
		ChannelSecret: plainSecret,
		BotTokenSet:   client.BotToken != "",
		Description:   client.Description,
		Status:        client.Status,
	}
	if client.BotToken != "" {
		plainToken, decErr := crypto.Decrypt(s.encKey, client.BotToken)
		if decErr == nil {
			resp.BotToken = maskBotToken(plainToken)
		}
	}
	return resp, nil
}

// DeleteChannelClient 删除渠道客户端（软删除）
func (s *ChannelClientService) DeleteChannelClient(id uint) error {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrChannelClientNotFound
	}
	return s.repo.Delete(client)
}

// UpdateChannelClientStatus 更新渠道客户端状态
func (s *ChannelClientService) UpdateChannelClientStatus(id uint, status int) error {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrChannelClientNotFound
	}
	client.Status = status
	return s.repo.Update(client)
}

// UpdateChannelClient 更新渠道客户端信息（名称、描述、bot_token）
func (s *ChannelClientService) UpdateChannelClient(id uint, name, description string, botToken *string) (*ChannelClientResponse, error) {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrChannelClientNotFound
	}

	if name != "" {
		client.Name = name
	}
	client.Description = description

	// botToken 为 nil 表示不修改；非 nil 则更新（空字符串表示清空）
	if botToken != nil {
		if *botToken != "" {
			encryptedToken, err := crypto.Encrypt(s.encKey, *botToken)
			if err != nil {
				return nil, fmt.Errorf("encrypt bot token: %w", err)
			}
			client.BotToken = encryptedToken
		} else {
			client.BotToken = ""
		}
	}

	if err := s.repo.Update(client); err != nil {
		return nil, err
	}

	plainSecret, err := crypto.Decrypt(s.encKey, client.ChannelSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt channel secret: %w", err)
	}

	resp := &ChannelClientResponse{
		ID:            client.ID,
		Name:          client.Name,
		ChannelType:   client.ChannelType,
		ChannelKey:    client.ChannelKey,
		ChannelSecret: plainSecret,
		BotTokenSet:   client.BotToken != "",
		Description:   client.Description,
		Status:        client.Status,
	}
	if client.BotToken != "" {
		plainToken, decErr := crypto.Decrypt(s.encKey, client.BotToken)
		if decErr == nil {
			resp.BotToken = maskBotToken(plainToken)
		}
	}
	return resp, nil
}

// DecryptBotToken 解密渠道客户端的 Bot Token（供 Channel API 使用）
func (s *ChannelClientService) DecryptBotToken(client *models.ChannelClient) (string, error) {
	if client.BotToken == "" {
		return "", nil
	}
	return crypto.Decrypt(s.encKey, client.BotToken)
}

// VerifyChannelSignature 验证渠道签名
// 复用 upstream/signer.go 的 HMAC-SHA256 签名算法
func (s *ChannelClientService) VerifyChannelSignature(key, signature string, timestamp int64, method, path string, body []byte) (*models.ChannelClient, error) {
	// 验证时间戳
	if !upstream.IsTimestampValid(timestamp) {
		return nil, ErrChannelTimestampExpired
	}

	// 查找客户端
	client, err := s.repo.FindByChannelKey(key)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrChannelClientNotFound
	}
	if client.Status != 1 {
		return nil, ErrChannelClientDisabled
	}

	// 解密 secret
	plainSecret, err := crypto.Decrypt(s.encKey, client.ChannelSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt channel secret: %w", err)
	}

	// 验证签名（复用 upstream.Verify）
	if !upstream.Verify(plainSecret, method, path, signature, timestamp, body) {
		return nil, ErrChannelSignatureInvalid
	}

	return client, nil
}
