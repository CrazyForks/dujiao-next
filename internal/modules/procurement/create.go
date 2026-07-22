package procurement

import (
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/queue"

	"github.com/google/uuid"
)

// CreateForOrder 为已支付订单创建采购单（上游交付类型）
func (s *Service) CreateForOrder(orderID uint) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("load order: %w", err)
	}
	if order == nil {
		return ErrOrderNotFound
	}

	// 父订单有子订单：遍历子订单
	if order.ParentID == nil && len(order.Children) > 0 {
		for i := range order.Children {
			child := &order.Children[i]
			if !s.hasUpstreamItems(child) {
				continue
			}
			if err := s.createProcurementForSingleOrder(child); err != nil {
				logger.Warnw("procurement_create_child_failed",
					"parent_order_id", orderID,
					"child_order_id", child.ID,
					"error", err,
				)
				return err
			}
		}
		return nil
	}

	// 单订单
	if !s.hasUpstreamItems(order) {
		return nil
	}
	return s.createProcurementForSingleOrder(order)
}

// createProcurementForSingleOrder 为单个订单创建采购单
func (s *Service) createProcurementForSingleOrder(order *models.Order) error {
	// 检查是否已存在
	existing, err := s.procRepo.GetByLocalOrderID(order.ID)
	if err != nil {
		return fmt.Errorf("check existing procurement: %w", err)
	}
	if existing != nil {
		return ErrExists
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("order %d has no items", order.ID)
	}
	item := order.Items[0]

	// 查找商品映射
	mapping, err := s.mappingRepo.GetByLocalProductID(item.ProductID)
	if err != nil {
		return fmt.Errorf("lookup product mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no product mapping for product %d", item.ProductID)
	}

	procOrder := &models.ProcurementOrder{
		ConnectionID:    mapping.ConnectionID,
		LocalOrderID:    order.ID,
		LocalOrderNo:    order.OrderNo,
		Status:          "pending",
		LocalSellAmount: order.TotalAmount,
		Currency:        order.Currency,
		TraceID:         uuid.NewString(),
	}

	if err := s.procRepo.Create(procOrder); err != nil {
		return fmt.Errorf("create procurement order: %w", err)
	}

	logger.Infow("procurement_order_created",
		"procurement_order_id", procOrder.ID,
		"local_order_id", order.ID,
		"connection_id", mapping.ConnectionID,
	)

	// 入队提交任务
	if s.queue != nil {
		if err := s.queue.EnqueueProcurementSubmit(queue.ProcurementSubmitPayload{
			ProcurementOrderID: procOrder.ID,
		}); err != nil {
			logger.Warnw("procurement_enqueue_submit_failed",
				"procurement_order_id", procOrder.ID,
				"error", err,
			)
		}
	}

	return nil
}

// hasUpstreamItems 检查订单是否包含上游交付类型的商品
func (s *Service) hasUpstreamItems(order *models.Order) bool {
	for _, item := range order.Items {
		if strings.TrimSpace(item.FulfillmentType) == constants.FulfillmentTypeUpstream {
			return true
		}
	}
	return false
}
