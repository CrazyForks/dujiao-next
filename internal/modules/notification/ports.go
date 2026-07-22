package notification

import (
	"context"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/dashboard"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/queue"

	"github.com/hibiken/asynq"
)

type NotificationCenterSetting = settingsmodule.NotificationCenterSetting
type DashboardSetting = settingsmodule.DashboardSetting
type DashboardAlertSetting = settingsmodule.DashboardAlertSetting

func normalizeNotificationLocale(locale string) string {
	return settingsmodule.NormalizeNotificationLocale(locale)
}

func normalizeNotificationInventoryAlertInterval(seconds int) int {
	return settingsmodule.NormalizeNotificationInventoryAlertInterval(seconds)
}

func normalizeNotificationPaymentOrderAlertInterval(seconds int) int {
	return settingsmodule.NormalizeNotificationPaymentOrderAlertInterval(seconds)
}

type SettingsReader interface {
	GetNotificationCenterSetting() (settingsmodule.NotificationCenterSetting, error)
	GetDashboardSetting() (settingsmodule.DashboardSetting, error)
}

type EmailSender interface {
	SendCustomEmail(toEmail, subject, body string) error
}

type Enqueuer interface {
	EnqueueNotificationDispatch(payload queue.NotificationDispatchPayload, opts ...asynq.Option) error
}

type DashboardAlertReader interface {
	LoadDashboardAlertSetting() settingsmodule.DashboardAlertSetting
	GetInventoryAlertItems(ctx context.Context, lowStockThreshold int64) ([]dashboard.InventoryAlertRow, error)
	GetPaymentOrderAlertCounts(ctx context.Context, startAt, endAt time.Time) (dashboard.PaymentOrderAlertCountsRow, error)
}

type TelegramSender interface {
	SendMessage(ctx context.Context, chatID, message string) error
}

type LogRepository interface {
	Create(log *models.NotificationLog) error
	ListAdmin(filter LogListFilter) ([]models.NotificationLog, int64, error)
}

type LogListFilter struct {
	Page        int
	PageSize    int
	Channel     string
	Status      string
	EventType   string
	IsTest      *bool
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}
