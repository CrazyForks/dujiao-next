package notification

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/queue"
)

type notificationSettingsStub struct {
	notification settingsmodule.NotificationCenterSetting
	dashboard    settingsmodule.DashboardSetting
}

func (s notificationSettingsStub) GetNotificationCenterSetting() (settingsmodule.NotificationCenterSetting, error) {
	return s.notification, nil
}

func (s notificationSettingsStub) GetDashboardSetting() (settingsmodule.DashboardSetting, error) {
	return s.dashboard, nil
}

type notificationEmailStub struct{}

func (notificationEmailStub) SendCustomEmail(recipient, _, _ string) error {
	if strings.Contains(recipient, "failure") {
		return errors.New("simulated email failure")
	}
	return nil
}

type notificationLogRepositoryStub struct {
	items []models.NotificationLog
}

func (r *notificationLogRepositoryStub) Create(item *models.NotificationLog) error {
	if item != nil {
		copy := *item
		copy.ID = uint(len(r.items) + 1)
		r.items = append(r.items, copy)
	}
	return nil
}

func (r *notificationLogRepositoryStub) ListAdmin(filter LogListFilter) ([]models.NotificationLog, int64, error) {
	result := make([]models.NotificationLog, 0, len(r.items))
	for _, item := range r.items {
		if filter.EventType != "" && item.EventType != filter.EventType {
			continue
		}
		if filter.IsTest != nil && item.IsTest != *filter.IsTest {
			continue
		}
		result = append(result, item)
	}
	return result, int64(len(result)), nil
}

func setupLogService(t *testing.T) (*Service, *LogService) {
	t.Helper()
	logService := NewLogService(&notificationLogRepositoryStub{})
	setting := settingsmodule.NotificationCenterSetting{
		DefaultLocale: constants.LocaleEnUS,
		Scenes: settingsmodule.NotificationSceneSetting{
			OrderPaidSuccess: true,
			ExceptionAlert:   true,
		},
		Channels: settingsmodule.NotificationChannelsSetting{
			Email: settingsmodule.NotificationChannelSetting{
				Enabled:    true,
				Recipients: []string{"success@example.com", "failure@example.com"},
			},
		},
		Templates: settingsmodule.NotificationTemplatesSetting{
			OrderPaidSuccess: settingsmodule.NotificationSceneTemplate{
				ENUS: settingsmodule.NotificationLocalizedTemplate{
					Title: "Order {{order_no}}",
					Body:  "Customer {{customer_email}}",
				},
			},
			ExceptionAlert: settingsmodule.NotificationSceneTemplate{
				ENUS: settingsmodule.NotificationLocalizedTemplate{
					Title: "Alert {{message}}",
					Body:  "{{message}}",
				},
			},
		},
	}
	service := NewService(notificationSettingsStub{notification: setting}, notificationEmailStub{}, nil, nil, logService, nil)
	return service, logService
}

func TestServiceSendTestRecordsSuccessLog(t *testing.T) {
	service, logService := setupLogService(t)
	if err := service.SendTest(context.Background(), TestSendInput{
		Channel: "email",
		Target:  "success@example.com",
		Scene:   constants.NotificationEventOrderPaidSuccess,
		Locale:  constants.LocaleEnUS,
	}); err != nil {
		t.Fatalf("SendTest failed: %v", err)
	}

	isTest := true
	items, total, err := logService.ListForAdmin(LogListFilter{
		Page: 1, PageSize: 10, EventType: constants.NotificationEventOrderPaidSuccess, IsTest: &isTest,
	})
	if err != nil {
		t.Fatalf("list notification logs failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 test log, total=%d len=%d", total, len(items))
	}
	if items[0].Status != notificationLogStatusSuccess || !items[0].IsTest {
		t.Fatalf("unexpected log: %#v", items[0])
	}
	if !strings.Contains(items[0].Title, "DJ202603230001") {
		t.Fatalf("title should include sample order no, got %s", items[0].Title)
	}
}

func TestServiceDispatchSingleEventRecordsPerRecipientResult(t *testing.T) {
	service, logService := setupLogService(t)
	setting, err := service.settingService.GetNotificationCenterSetting()
	if err != nil {
		t.Fatalf("get notification center setting failed: %v", err)
	}
	dispatchErr := service.dispatchSingleEvent(context.Background(), setting, queue.NotificationDispatchPayload{
		EventType: constants.NotificationEventOrderPaidSuccess,
		BizType:   constants.NotificationBizTypeOrder,
		BizID:     88,
		Locale:    constants.LocaleEnUS,
		Force:     true,
		Data: map[string]interface{}{
			"order_no":       "DJ-LOG-88",
			"customer_email": "member@example.com",
		},
	})
	if !errors.Is(dispatchErr, ErrSendFailed) {
		t.Fatalf("expected send failure, got %v", dispatchErr)
	}

	items, total, err := logService.ListForAdmin(LogListFilter{Page: 1, PageSize: 10, EventType: constants.NotificationEventOrderPaidSuccess})
	if err != nil {
		t.Fatalf("list notification logs failed: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 recipient logs, total=%d len=%d", total, len(items))
	}
	statuses := map[string]string{}
	for _, item := range items {
		statuses[item.Recipient] = item.Status
	}
	if statuses["success@example.com"] != notificationLogStatusSuccess {
		t.Fatalf("success recipient status mismatch: %v", statuses)
	}
	if statuses["failure@example.com"] != notificationLogStatusFailed {
		t.Fatalf("failure recipient status mismatch: %v", statuses)
	}
}
