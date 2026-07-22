package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dujiao-next/internal/config"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
)

type telegramSendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// SendOptions Telegram 发送参数。
type SendOptions struct {
	ChatID                string
	Message               string
	ParseMode             string
	DisableWebPagePreview bool
	AttachmentURL         string
	AttachmentDisplayName string
}

// SettingReader 是通知发送器读取 Telegram 动态配置所需的最小端口。
type SettingReader interface {
	GetTelegramAuthSetting(defaultCfg config.TelegramAuthConfig) (settingsmodule.TelegramAuthSetting, error)
}

// NotifyService Telegram 通知发送服务。
type NotifyService struct {
	settingService SettingReader
	defaultCfg     config.TelegramAuthConfig
	httpClient     *http.Client
}

// NewNotifyService 创建 Telegram 通知发送服务。
func NewNotifyService(settingService SettingReader, defaultCfg config.TelegramAuthConfig) *NotifyService {
	return &NotifyService{
		settingService: settingService,
		defaultCfg:     defaultCfg,
		httpClient: &http.Client{
			Timeout: 6 * time.Second,
		},
	}
}

// SendMessage 发送 Telegram 消息
func (s *NotifyService) SendMessage(ctx context.Context, chatID, message string) error {
	token, err := s.resolveBotToken()
	if err != nil {
		return err
	}
	if token == "" {
		return ErrNotifyConfigInvalid
	}
	return s.SendWithBotToken(ctx, token, SendOptions{
		ChatID:                chatID,
		Message:               message,
		DisableWebPagePreview: true,
	})
}

// SendWithBotToken 使用显式 bot token 发送 Telegram 消息。
func (s *NotifyService) SendWithBotToken(ctx context.Context, botToken string, options SendOptions) error {
	chatID := strings.TrimSpace(options.ChatID)
	message := strings.TrimSpace(options.Message)
	botToken = strings.TrimSpace(botToken)
	if chatID == "" || message == "" || botToken == "" {
		return ErrNotifySendFailed
	}

	if strings.TrimSpace(options.AttachmentURL) != "" {
		attachmentURL := strings.TrimSpace(options.AttachmentURL)
		if isTelegramPhotoAttachment(attachmentURL, options.AttachmentDisplayName) {
			if filePath, ok := resolveTelegramAttachmentPath(attachmentURL); ok {
				return s.sendMultipartMedia(ctx, botToken, "sendPhoto", "photo", filePath, options)
			}
			payload := map[string]interface{}{
				"chat_id": chatID,
				"photo":   attachmentURL,
				"caption": message,
			}
			if parseMode := strings.TrimSpace(options.ParseMode); parseMode != "" {
				payload["parse_mode"] = parseMode
			}
			return s.sendJSONRequest(ctx, botToken, "sendPhoto", payload)
		}
		if filePath, ok := resolveTelegramAttachmentPath(attachmentURL); ok {
			return s.sendMultipartMedia(ctx, botToken, "sendDocument", "document", filePath, options)
		}
		payload := map[string]interface{}{
			"chat_id":  chatID,
			"document": attachmentURL,
			"caption":  message,
		}
		if parseMode := strings.TrimSpace(options.ParseMode); parseMode != "" {
			payload["parse_mode"] = parseMode
		}
		return s.sendJSONRequest(ctx, botToken, "sendDocument", payload)
	}

	payload := map[string]interface{}{
		"chat_id":                  chatID,
		"text":                     message,
		"disable_web_page_preview": options.DisableWebPagePreview,
	}
	if parseMode := strings.TrimSpace(options.ParseMode); parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	return s.sendJSONRequest(ctx, botToken, "sendMessage", payload)
}

func (s *NotifyService) sendMultipartMedia(ctx context.Context, botToken, method, fieldName, filePath string, options SendOptions) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("%w: open attachment failed: %v", ErrNotifySendFailed, err)
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", strings.TrimSpace(options.ChatID)); err != nil {
		return err
	}
	if caption := strings.TrimSpace(options.Message); caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}
	if parseMode := strings.TrimSpace(options.ParseMode); parseMode != "" {
		if err := writer.WriteField("parse_mode", parseMode); err != nil {
			return err
		}
	}
	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	requestURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return s.doRequest(req)
}

func (s *NotifyService) sendJSONRequest(ctx context.Context, botToken, method string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return s.doRequest(req)
}

func (s *NotifyService) doRequest(req *http.Request) error {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotifySendFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotifySendFailed, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: telegram status=%d body=%s", ErrNotifySendFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed telegramSendMessageResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("%w: parse telegram response failed", ErrNotifySendFailed)
	}
	if !parsed.OK {
		return fmt.Errorf("%w: %s", ErrNotifySendFailed, strings.TrimSpace(parsed.Description))
	}
	return nil
}

func resolveTelegramAttachmentPath(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", false
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed != nil && parsed.Scheme != "" {
		return "", false
	}

	normalized := strings.TrimPrefix(value, "/")
	normalized = filepath.Clean(normalized)
	if normalized == "." || normalized == "" {
		return "", false
	}
	if normalized == "uploads" || strings.HasPrefix(normalized, "uploads"+string(filepath.Separator)) {
		return normalized, true
	}
	return "", false
}

func isTelegramPhotoAttachment(rawURL, displayName string) bool {
	candidates := []string{
		strings.TrimSpace(displayName),
		strings.TrimSpace(rawURL),
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		value := candidate
		if parsed, err := url.Parse(candidate); err == nil && parsed != nil {
			if parsed.Path != "" {
				value = parsed.Path
			}
		}
		ext := strings.ToLower(strings.TrimSpace(filepath.Ext(value)))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
			return true
		}
		if ext == ".gif" {
			return true
		}
		if detected := mime.TypeByExtension(ext); strings.HasPrefix(strings.ToLower(detected), "image/") {
			return true
		}
	}

	return false
}

func (s *NotifyService) resolveBotToken() (string, error) {
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
