package notification

import (
	"context"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/queue"

	"github.com/hibiken/asynq"
)

// EnqueueInput 通知事件入队参数。
type EnqueueInput struct {
	EventType string
	BizType   string
	BizID     uint
	Locale    string
	Force     bool
	Data      models.JSON
}

// Service 通知中心服务。
type Service struct {
	settingService SettingsReader
	emailService   EmailSender
	queueClient    Enqueuer
	dashboardSvc   DashboardAlertReader
	logService     *LogService
	telegramSender TelegramSender
}

// NewService 创建通知中心服务。
func NewService(
	settingService SettingsReader,
	emailService EmailSender,
	queueClient Enqueuer,
	dashboardSvc DashboardAlertReader,
	logService *LogService,
	telegramSender TelegramSender,
) *Service {
	return &Service{
		settingService: settingService,
		emailService:   emailService,
		queueClient:    queueClient,
		dashboardSvc:   dashboardSvc,
		logService:     logService,
		telegramSender: telegramSender,
	}
}

// Enqueue 入队通知任务
func (s *Service) Enqueue(input EnqueueInput) error {
	eventType := strings.ToLower(strings.TrimSpace(input.EventType))
	if !isNotificationEventSupported(eventType) {
		return ErrEventInvalid
	}
	if s == nil || s.queueClient == nil {
		return nil
	}

	payload := queue.NotificationDispatchPayload{
		EventType: eventType,
		BizType:   strings.TrimSpace(input.BizType),
		BizID:     input.BizID,
		Locale:    strings.TrimSpace(input.Locale),
		Force:     input.Force,
		Data:      notificationJSONToMap(input.Data),
	}
	return s.queueClient.EnqueueNotificationDispatch(payload, asynq.MaxRetry(5))
}

// Dispatch 处理通知分发任务
func (s *Service) Dispatch(ctx context.Context, payload queue.NotificationDispatchPayload) error {
	if s == nil {
		return nil
	}
	eventType := strings.ToLower(strings.TrimSpace(payload.EventType))
	if !isNotificationEventSupported(eventType) {
		return ErrEventInvalid
	}

	setting, err := s.settingService.GetNotificationCenterSetting()
	if err != nil {
		return err
	}
	if !setting.Scenes.IsSceneEnabled(eventType) {
		return nil
	}

	if eventType == constants.NotificationEventExceptionAlertCheck {
		return s.dispatchExceptionAlertCheck(ctx, setting, payload)
	}
	return s.dispatchSingleEvent(ctx, setting, payload)
}
