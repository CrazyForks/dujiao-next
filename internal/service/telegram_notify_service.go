package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dujiao-next/internal/config"
)

type telegramSendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// TelegramNotifyService Telegram 通知发送服务
type TelegramNotifyService struct {
	settingService *SettingService
	defaultCfg     config.TelegramAuthConfig
	httpClient     *http.Client
}

// NewTelegramNotifyService 创建 Telegram 通知发送服务
func NewTelegramNotifyService(settingService *SettingService, defaultCfg config.TelegramAuthConfig) *TelegramNotifyService {
	return &TelegramNotifyService{
		settingService: settingService,
		defaultCfg:     defaultCfg,
		httpClient: &http.Client{
			Timeout: 6 * time.Second,
		},
	}
}

// SendMessage 发送 Telegram 消息
func (s *TelegramNotifyService) SendMessage(ctx context.Context, chatID, message string) error {
	chatID = strings.TrimSpace(chatID)
	message = strings.TrimSpace(message)
	if chatID == "" || message == "" {
		return ErrNotificationSendFailed
	}
	token, err := s.resolveBotToken()
	if err != nil {
		return err
	}
	if token == "" {
		return ErrNotificationConfigInvalid
	}

	payloadMap := map[string]interface{}{
		"chat_id":                  chatID,
		"text":                     message,
		"disable_web_page_preview": true,
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotificationSendFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotificationSendFailed, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: telegram status=%d body=%s", ErrNotificationSendFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed telegramSendMessageResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("%w: parse telegram response failed", ErrNotificationSendFailed)
	}
	if !parsed.OK {
		return fmt.Errorf("%w: %s", ErrNotificationSendFailed, strings.TrimSpace(parsed.Description))
	}
	return nil
}

func (s *TelegramNotifyService) resolveBotToken() (string, error) {
	if s == nil {
		return "", nil
	}
	if s.settingService == nil {
		return strings.TrimSpace(s.defaultCfg.BotToken), nil
	}
	setting, err := s.settingService.GetTelegramAuthSetting(s.defaultCfg)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(setting.BotToken), nil
}
