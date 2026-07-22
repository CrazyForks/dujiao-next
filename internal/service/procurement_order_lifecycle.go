package service

import (
	"time"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/procurement"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/upstream"

	"gorm.io/gorm"
)

// procurementOrderLifecycleAdapter 把仍属于订单领域的事务与通知策略适配给采购模块。
// 采购模块因此无需依赖 service 包中的具体服务，也不会复制订单状态聚合规则。
type procurementOrderLifecycleAdapter struct {
	orders             repository.OrderRepository
	fulfillments       repository.FulfillmentRepository
	queue              *queue.Client
	settings           *settingsapp.Service
	defaultEmailConfig config.EmailConfig
}

func NewProcurementOrderLifecycle(
	orders repository.OrderRepository,
	fulfillments repository.FulfillmentRepository,
	queueClient *queue.Client,
	settings *settingsapp.Service,
	defaultEmailConfig config.EmailConfig,
) procurement.OrderLifecycle {
	return &procurementOrderLifecycleAdapter{
		orders:             orders,
		fulfillments:       fulfillments,
		queue:              queueClient,
		settings:           settings,
		defaultEmailConfig: defaultEmailConfig,
	}
}

func (a *procurementOrderLifecycleAdapter) CreateUpstreamFulfillment(
	orderID uint,
	upstreamFulfillment *upstream.UpstreamFulfillment,
	now time.Time,
) error {
	deliveredAt := upstreamFulfillment.DeliveredAt
	if deliveredAt == nil {
		deliveredAt = &now
	}

	return a.orders.Transaction(func(tx *gorm.DB) error {
		fulfillments := a.fulfillments.WithTx(tx)
		if _, found, err := fulfillments.FindByOrderIDForUpdate(orderID); err != nil {
			return err
		} else if found {
			return nil
		}

		return fulfillments.Create(&models.Fulfillment{
			OrderID:       orderID,
			Type:          constants.FulfillmentTypeUpstream,
			Status:        constants.FulfillmentStatusDelivered,
			Payload:       upstreamFulfillment.Payload,
			LogisticsJSON: upstreamFulfillment.DeliveryData,
			DeliveredAt:   deliveredAt,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	})
}

func (a *procurementOrderLifecycleAdapter) SyncParentStatus(parentID uint, now time.Time) (string, error) {
	return syncParentStatus(a.orders, parentID, now)
}

func (a *procurementOrderLifecycleAdapter) EnqueueStatusEmail(orderID uint, status string) (bool, error) {
	return enqueueOrderStatusEmailTaskIfEligible(
		a.orders,
		a.queue,
		a.settings,
		a.defaultEmailConfig,
		orderID,
		status,
	)
}
