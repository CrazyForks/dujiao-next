package procurement

import (
	"context"
	"fmt"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/queue"
)

// RetryManual 手动重试失败的采购单
func (s *Service) RetryManual(id uint) error {
	procOrder, err := s.procRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("load procurement order: %w", err)
	}
	if procOrder == nil {
		return ErrNotFound
	}

	if procOrder.Status != "failed" && procOrder.Status != "rejected" {
		return ErrStatusInvalid
	}

	now := time.Now()
	updates := map[string]interface{}{
		"retry_count":   0,
		"next_retry_at": nil,
		"error_message": "",
		"updated_at":    now,
	}
	if err := s.procRepo.UpdateStatus(procOrder.ID, "pending", updates); err != nil {
		return fmt.Errorf("reset procurement status: %w", err)
	}

	logger.Infow("procurement_manual_retry",
		"procurement_order_id", procOrder.ID,
	)

	if s.queue != nil {
		return s.queue.EnqueueProcurementSubmit(queue.ProcurementSubmitPayload{
			ProcurementOrderID: procOrder.ID,
		})
	}
	return nil
}

// CancelManual 手动取消采购单
func (s *Service) CancelManual(id uint) error {
	procOrder, err := s.procRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("load procurement order: %w", err)
	}
	if procOrder == nil {
		return ErrNotFound
	}

	// 已交付/已退款的不能取消
	if procOrder.Status == constants.ProcurementStatusFulfilled ||
		procOrder.Status == constants.ProcurementStatusCompleted ||
		procOrder.Status == constants.ProcurementStatusPartiallyRefunded ||
		procOrder.Status == constants.ProcurementStatusRefunded ||
		procOrder.Status == constants.ProcurementStatusCanceled {
		return ErrStatusInvalid
	}

	// 已被上游接受：尝试取消上游订单
	if procOrder.Status == "accepted" && procOrder.UpstreamOrderID > 0 {
		conn, err := s.connections.GetByID(procOrder.ConnectionID)
		if err == nil && conn != nil {
			adapter, adErr := s.connections.GetAdapter(conn)
			if adErr == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				if cancelErr := adapter.CancelOrder(ctx, procOrder.UpstreamOrderID); cancelErr != nil {
					logger.Warnw("procurement_cancel_upstream_failed",
						"procurement_order_id", procOrder.ID,
						"upstream_order_id", procOrder.UpstreamOrderID,
						"error", cancelErr,
					)
				}
			}
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"error_message": "manually canceled",
		"updated_at":    now,
	}
	if err := s.procRepo.UpdateStatus(procOrder.ID, "canceled", updates); err != nil {
		return fmt.Errorf("update procurement status: %w", err)
	}

	logger.Infow("procurement_manual_cancel",
		"procurement_order_id", procOrder.ID,
	)
	return nil
}
