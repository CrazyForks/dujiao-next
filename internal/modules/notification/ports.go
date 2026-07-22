package notification

import (
	"context"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/dashboard"
	settingsmessaging "github.com/dujiao-next/internal/modules/settings/schema/messaging"
	settingsstorefront "github.com/dujiao-next/internal/modules/settings/schema/storefront"
	"github.com/dujiao-next/internal/queue"

	"github.com/hibiken/asynq"
)

type NotificationCenterSetting = settingsmessaging.NotificationCenterSetting
type DashboardSetting = settingsstorefront.DashboardSetting
type DashboardAlertSetting = settingsstorefront.DashboardAlertSetting

func normalizeNotificationLocale(locale string) string {
	return settingsmessaging.NormalizeNotificationLocale(locale)
}

func normalizeNotificationInventoryAlertInterval(seconds int) int {
	return settingsmessaging.NormalizeNotificationInventoryAlertInterval(seconds)
}

func normalizeNotificationPaymentOrderAlertInterval(seconds int) int {
	return settingsmessaging.NormalizeNotificationPaymentOrderAlertInterval(seconds)
}

type SettingsReader interface {
	GetNotificationCenterSetting() (settingsmessaging.NotificationCenterSetting, error)
	GetDashboardSetting() (settingsstorefront.DashboardSetting, error)
}

type EmailSender interface {
	SendCustomEmail(toEmail, subject, body string) error
}

type Enqueuer interface {
	EnqueueNotificationDispatch(payload queue.NotificationDispatchPayload, opts ...asynq.Option) error
}

type DashboardAlertReader interface {
	LoadDashboardAlertSetting() settingsstorefront.DashboardAlertSetting
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
