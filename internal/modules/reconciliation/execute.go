package reconciliation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	notificationcontract "github.com/dujiao-next/internal/modules/notification/contract"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/dujiao-next/internal/upstream"

	"github.com/shopspring/decimal"
)

func (s *Service) execute(ctx context.Context, job *models.ReconciliationJob) error {
	connection, err := s.connections.GetByID(job.ConnectionID)
	if err != nil {
		return fmt.Errorf("get connection: %w", err)
	}
	adapter, err := s.connections.GetAdapter(connection)
	if err != nil {
		return fmt.Errorf("get adapter: %w", err)
	}
	orders, err := s.procurements.ListByConnectionAndTimeRange(job.ConnectionID, job.TimeRangeStart, job.TimeRangeEnd)
	if err != nil {
		return fmt.Errorf("list procurement orders: %w", err)
	}

	mismatches := make([]models.ReconciliationItem, 0)
	skippedCount, errorCount := 0, 0
	for index := range orders {
		order := &orders[index]
		if order.UpstreamOrderID == 0 {
			skippedCount++
			continue
		}
		detail, err := adapter.GetOrder(ctx, order.UpstreamOrderID)
		if err != nil {
			logger.Warnw("reconciliation_get_upstream_order_failed", "job_id", job.ID,
				"procurement_id", order.ID, "upstream_order_id", order.UpstreamOrderID, "error", err)
			errorCount++
			continue
		}
		if item := compareOrder(job, order, detail); item != nil {
			mismatches = append(mismatches, *item)
		}
	}
	if len(mismatches) > 0 {
		if err := s.items.BatchCreate(mismatches); err != nil {
			return fmt.Errorf("batch create reconciliation items: %w", err)
		}
	}

	comparedCount := len(orders) - skippedCount - errorCount
	job.TotalCount = comparedCount
	job.MismatchedCount = len(mismatches)
	job.MatchedCount = comparedCount - job.MismatchedCount
	job.ResultJSON = marshalResult(map[string]any{
		"total": job.TotalCount, "matched": job.MatchedCount, "mismatched": job.MismatchedCount,
		"skipped": skippedCount, "errors": errorCount,
	})
	return nil
}

func compareOrder(job *models.ReconciliationJob, order *models.ProcurementOrder, detail *upstream.UpstreamOrderDetail) *models.ReconciliationItem {
	checkStatus := job.Type == constants.ReconciliationTypeStatus || job.Type == constants.ReconciliationTypeFull
	checkAmount := job.Type == constants.ReconciliationTypeAmount || job.Type == constants.ReconciliationTypeFull
	statusMismatch := checkStatus && !isStatusConsistent(order.Status, detail.Status)

	amountMismatch := false
	var upstreamAmount money.Amount
	if checkAmount && detail.Amount != "" {
		value, err := decimal.NewFromString(detail.Amount)
		if err == nil && value.IsPositive() && order.UpstreamAmount.IsPositive() {
			upstreamAmount = money.FromDecimal(value)
			amountMismatch = !order.UpstreamAmount.Equal(value)
		}
	}

	mismatchType := ""
	switch {
	case statusMismatch && amountMismatch:
		mismatchType = constants.MismatchTypeBoth
	case statusMismatch:
		mismatchType = constants.MismatchTypeStatus
	case amountMismatch:
		mismatchType = constants.MismatchTypeAmount
	}
	if mismatchType == "" {
		return nil
	}
	return &models.ReconciliationItem{
		JobID: job.ID, ProcurementOrderID: order.ID,
		LocalOrderNo: order.LocalOrderNo, UpstreamOrderNo: order.UpstreamOrderNo,
		LocalStatus: order.Status, UpstreamStatus: detail.Status,
		LocalAmount: order.UpstreamAmount, UpstreamAmount: upstreamAmount, MismatchType: mismatchType,
	}
}

func isStatusConsistent(localStatus, upstreamStatus string) bool {
	localStatus = strings.ToLower(strings.TrimSpace(localStatus))
	upstreamStatus = strings.ToLower(strings.TrimSpace(upstreamStatus))
	switch localStatus {
	case constants.ProcurementStatusCompleted, constants.ProcurementStatusFulfilled:
		return upstreamStatus == "completed" || upstreamStatus == "delivered" || upstreamStatus == "fulfilled" ||
			upstreamStatus == "refunded" || upstreamStatus == "partially_refunded"
	case constants.ProcurementStatusCanceled:
		return upstreamStatus == "canceled" || upstreamStatus == "cancelled" || upstreamStatus == "refunded" || upstreamStatus == "partially_refunded"
	case constants.ProcurementStatusPending:
		return upstreamStatus == "pending" || upstreamStatus == "paid"
	case constants.ProcurementStatusSubmitted, constants.ProcurementStatusAccepted:
		return upstreamStatus == "paid" || upstreamStatus == "processing" || upstreamStatus == "accepted"
	case constants.ProcurementStatusFailed, constants.ProcurementStatusRejected:
		return upstreamStatus == "failed" || upstreamStatus == "rejected"
	case "fulfilling":
		return upstreamStatus == "fulfilling" || upstreamStatus == "processing" || upstreamStatus == "paid"
	default:
		return localStatus == upstreamStatus
	}
}

func marshalResult(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func (s *Service) notifyMismatch(job *models.ReconciliationJob) {
	if s.notifications == nil {
		return
	}
	_ = s.notifications.Enqueue(notificationcontract.EnqueueInput{
		EventType: constants.NotificationEventExceptionAlert,
		BizType:   constants.NotificationBizTypeReconciliation, BizID: job.ID,
		Data: map[string]any{
			"message": fmt.Sprintf("对账任务 #%d 完成，发现 %d 项差异", job.ID, job.MismatchedCount),
			"job_id":  job.ID, "connection_id": job.ConnectionID,
			"total_count": job.TotalCount, "mismatched_count": job.MismatchedCount,
		},
	})
}
