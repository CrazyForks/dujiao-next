package notification

import (
	"context"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/queue"
)

func (s *Service) dispatchExceptionAlertCheck(ctx context.Context, setting NotificationCenterSetting, payload queue.NotificationDispatchPayload) error {
	if s.dashboardSvc == nil || s.settingService == nil {
		return nil
	}

	dashboardSetting, err := s.settingService.GetDashboardSetting()
	if err != nil {
		return err
	}

	var firstErr error
	inventoryAlerts, err := s.dashboardSvc.GetInventoryAlertItems(ctx, dashboardSetting.Alert.LowStockThreshold)
	if err != nil {
		return err
	}
	for _, itemPayload := range BuildInventoryAlertDispatchPayloads(setting, dashboardSetting, payload, inventoryAlerts) {
		allowed, intervalErr := acquireInventoryAlertInterval(ctx, setting.InventoryAlertIntervalSeconds, itemPayload)
		if intervalErr != nil {
			logger.Warnw("notification_inventory_alert_interval_failed", "error", intervalErr)
		}
		if intervalErr == nil && !allowed {
			continue
		}
		if err := s.dispatchSingleEvent(ctx, setting, itemPayload); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	paymentOrderAlertNow := time.Now()
	paymentOrderAlertStart := paymentOrderAlertNow.Add(-time.Duration(setting.PaymentOrderAlertCheckSeconds) * time.Second)
	paymentOrderCounts, err := s.dashboardSvc.GetPaymentOrderAlertCounts(ctx, paymentOrderAlertStart, paymentOrderAlertNow)
	if err != nil {
		return err
	}

	for _, itemPayload := range BuildPaymentOrderAlertDispatchPayloads(setting, dashboardSetting, payload, paymentOrderCounts) {
		allowed, intervalErr := acquirePaymentOrderAlertInterval(ctx, setting.PaymentOrderAlertIntervalSeconds, itemPayload)
		if intervalErr != nil {
			logger.Warnw("notification_payment_order_alert_interval_failed", "error", intervalErr)
		}
		if intervalErr == nil && !allowed {
			continue
		}
		if err := s.dispatchSingleEvent(ctx, setting, itemPayload); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func thresholdValueByAlertType(setting DashboardAlertSetting, alertType string) int64 {
	switch alertType {
	case constants.NotificationAlertTypeOutOfStockProducts:
		return setting.OutOfStockProductsThreshold
	case constants.NotificationAlertTypeLowStockProducts:
		return setting.LowStockThreshold
	case constants.NotificationAlertTypePendingOrders:
		return setting.PendingPaymentOrdersThreshold
	case constants.NotificationAlertTypePaymentsFailed:
		return setting.PaymentsFailedThreshold
	default:
		return 0
	}
}

func alertTypeLabelByType(locale, alertType string) string {
	type labels struct{ zhCN, zhTW, enUS string }
	m := map[string]labels{
		constants.NotificationAlertTypeOutOfStockProducts: {"售罄商品", "售罄商品", "Out of Stock"},
		constants.NotificationAlertTypeLowStockProducts:   {"低库存商品", "低庫存商品", "Low Stock"},
		constants.NotificationAlertTypePendingOrders:      {"待支付订单", "待支付訂單", "Pending Payment"},
		constants.NotificationAlertTypePaymentsFailed:     {"支付失败", "支付失敗", "Payment Failed"},
	}
	l, ok := m[alertType]
	if !ok {
		return alertType
	}
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "zh-tw":
		return l.zhTW
	case "en-us", "en":
		return l.enUS
	default:
		return l.zhCN
	}
}
