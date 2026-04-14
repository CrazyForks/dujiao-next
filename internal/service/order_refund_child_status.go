package service

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"

	"gorm.io/gorm"
)

// applyParentRefundChildStatusUpdatesTx 根据父订单退款结果统一更新子订单状态。
// 规则：子订单退款状态始终跟随父订单，避免分支逻辑造成状态混乱。
func applyParentRefundChildStatusUpdatesTx(tx *gorm.DB, parentOrderID uint, parentTargetStatus string, now time.Time) error {
	if tx == nil || parentOrderID == 0 {
		return nil
	}

	target := strings.ToLower(strings.TrimSpace(parentTargetStatus))
	if target != constants.OrderStatusPartiallyRefunded && target != constants.OrderStatusRefunded {
		return nil
	}

	return tx.Model(&models.Order{}).
		Where("parent_id = ? AND status <> ?", parentOrderID, target).
		Updates(map[string]interface{}{
			"status":     target,
			"updated_at": now,
		}).Error
}
