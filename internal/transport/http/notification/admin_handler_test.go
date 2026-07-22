package notificationhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/notification"

	"github.com/gin-gonic/gin"
)

type notificationLogServiceStub struct {
	items []models.NotificationLog
}

func (s notificationLogServiceStub) ListForAdmin(filter notification.LogListFilter) ([]models.NotificationLog, int64, error) {
	result := make([]models.NotificationLog, 0, len(s.items))
	for _, item := range s.items {
		if filter.Channel != "" && item.Channel != filter.Channel {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.IsTest != nil && item.IsTest != *filter.IsTest {
			continue
		}
		result = append(result, item)
	}
	return result, int64(len(result)), nil
}

func TestListNotificationLogsFiltersStatusAndChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC().Truncate(time.Second)

	items := []models.NotificationLog{
		{
			EventType:    "order_paid_success",
			BizType:      "order",
			BizID:        11,
			Channel:      "email",
			Recipient:    "failed@example.com",
			Locale:       "en-US",
			Title:        "Order DJ-11",
			Body:         "failed body",
			Status:       "failed",
			ErrorMessage: "smtp disabled",
			IsTest:       false,
			CreatedAt:    now,
		},
		{
			EventType: "order_paid_success",
			BizType:   "order",
			BizID:     12,
			Channel:   "telegram",
			Recipient: "-100100",
			Locale:    "zh-CN",
			Title:     "Telegram notice",
			Body:      "ok",
			Status:    "success",
			IsTest:    true,
			CreatedAt: now.Add(time.Second),
		},
	}
	h := &AdminHandler{logs: notificationLogServiceStub{items: items}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/settings/notification-center/logs?channel=email&status=failed&is_test=false", nil)

	h.ListNotificationLogs(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Data       []models.NotificationLog `json:"data"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if resp.Pagination.Total != 1 {
		t.Fatalf("pagination total want 1 got %d", resp.Pagination.Total)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("data len want 1 got %d", len(resp.Data))
	}
	if resp.Data[0].Recipient != "failed@example.com" {
		t.Fatalf("unexpected recipient: %s", resp.Data[0].Recipient)
	}
	if resp.Data[0].Status != "failed" {
		t.Fatalf("unexpected status: %s", resp.Data[0].Status)
	}
}
