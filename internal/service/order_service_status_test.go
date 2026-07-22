package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestCalcParentStatus(t *testing.T) {
	children := []models.Order{
		{Status: constants.OrderStatusDelivered},
		{Status: constants.OrderStatusPaid},
	}
	status := calcParentStatus(children, constants.OrderStatusPaid)
	if status != constants.OrderStatusPartiallyDelivered {
		t.Fatalf("expected partially_delivered, got %s", status)
	}

	children = []models.Order{
		{Status: constants.OrderStatusDelivered},
		{Status: constants.OrderStatusDelivered},
	}
	status = calcParentStatus(children, constants.OrderStatusPaid)
	if status != constants.OrderStatusDelivered {
		t.Fatalf("expected delivered, got %s", status)
	}
}

func TestCalcParentStatusAllRefunded(t *testing.T) {
	children := []models.Order{
		{Status: constants.OrderStatusRefunded},
		{Status: constants.OrderStatusRefunded},
	}
	status := calcParentStatus(children, constants.OrderStatusDelivered)
	if status != constants.OrderStatusRefunded {
		t.Fatalf("expected refunded, got %s", status)
	}
}

func TestCalcParentStatusPartiallyRefunded(t *testing.T) {
	children := []models.Order{
		{Status: constants.OrderStatusRefunded},
		{Status: constants.OrderStatusDelivered},
	}
	status := calcParentStatus(children, constants.OrderStatusDelivered)
	if status != constants.OrderStatusPartiallyRefunded {
		t.Fatalf("expected partially_refunded, got %s", status)
	}
}

func TestExpectedRefundStatus(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		order  models.Order
		expect string
	}{
		{
			name: "partial refund",
			order: models.Order{
				Status:         constants.OrderStatusCompleted,
				PaidAt:         &now,
				TotalAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
				RefundedAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(30)),
			},
			expect: constants.OrderStatusPartiallyRefunded,
		},
		{
			name: "full refund",
			order: models.Order{
				Status:         constants.OrderStatusCompleted,
				PaidAt:         &now,
				TotalAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
				RefundedAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
			},
			expect: constants.OrderStatusRefunded,
		},
		{
			name: "canceled should keep",
			order: models.Order{
				Status:         constants.OrderStatusCanceled,
				PaidAt:         &now,
				TotalAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
				RefundedAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
			},
			expect: "",
		},
	}

	for _, tc := range tests {
		got := expectedRefundStatus(&tc.order)
		if got != tc.expect {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.expect, got)
		}
	}
}

func TestResolvedParentStatusPrefersOwnRefund(t *testing.T) {
	now := time.Now()
	order := &models.Order{
		Status:         constants.OrderStatusCompleted,
		PaidAt:         &now,
		TotalAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(40)),
		RefundedAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		Children: []models.Order{
			{Status: constants.OrderStatusCompleted},
			{Status: constants.OrderStatusCompleted},
		},
	}
	if got := resolvedParentStatus(order); got != constants.OrderStatusPartiallyRefunded {
		t.Fatalf("expected partially_refunded, got %s", got)
	}
}

func TestIsTransitionAllowedRefunded(t *testing.T) {
	if !isTransitionAllowed(constants.OrderStatusDelivered, constants.OrderStatusPartiallyRefunded) {
		t.Fatalf("expected delivered to partially_refunded transition to be allowed")
	}
	if !isTransitionAllowed(constants.OrderStatusPartiallyRefunded, constants.OrderStatusRefunded) {
		t.Fatalf("expected partially_refunded to refunded transition to be allowed")
	}
	if !isTransitionAllowed(constants.OrderStatusDelivered, constants.OrderStatusRefunded) {
		t.Fatalf("expected delivered to refunded transition to be allowed")
	}
	if !isTransitionAllowed(constants.OrderStatusCompleted, constants.OrderStatusRefunded) {
		t.Fatalf("expected completed to refunded transition to be allowed")
	}
	if isTransitionAllowed(constants.OrderStatusCanceled, constants.OrderStatusRefunded) {
		t.Fatalf("expected canceled to refunded transition to be rejected")
	}
}

func TestUpdateOrderStatusParentToPartiallyRefundedSyncsChildren(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_parent_partial_refund_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}, &models.Fulfillment{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now()
	paidAt := now
	parent := &models.Order{
		OrderNo:          "PARENT-PARTIAL-REFUND-001",
		UserID:           0,
		Status:           constants.OrderStatusDelivered,
		Currency:         "CNY",
		OriginalAmount:   models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		DiscountAmount:   models.NewMoneyFromDecimal(decimal.Zero),
		TotalAmount:      models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		WalletPaidAmount: models.NewMoneyFromDecimal(decimal.Zero),
		OnlinePaidAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		RefundedAmount:   models.NewMoneyFromDecimal(decimal.NewFromInt(30)),
		PaidAt:           &paidAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.Create(parent).Error; err != nil {
		t.Fatalf("create parent order failed: %v", err)
	}

	childA := &models.Order{
		OrderNo:          "PARENT-PARTIAL-REFUND-001-A",
		ParentID:         &parent.ID,
		UserID:           0,
		Status:           constants.OrderStatusDelivered,
		Currency:         "CNY",
		OriginalAmount:   models.NewMoneyFromDecimal(decimal.NewFromInt(60)),
		DiscountAmount:   models.NewMoneyFromDecimal(decimal.Zero),
		TotalAmount:      models.NewMoneyFromDecimal(decimal.NewFromInt(60)),
		WalletPaidAmount: models.NewMoneyFromDecimal(decimal.Zero),
		OnlinePaidAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(60)),
		RefundedAmount:   models.NewMoneyFromDecimal(decimal.Zero),
		PaidAt:           &paidAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.Create(childA).Error; err != nil {
		t.Fatalf("create childA order failed: %v", err)
	}

	childB := &models.Order{
		OrderNo:          "PARENT-PARTIAL-REFUND-001-B",
		ParentID:         &parent.ID,
		UserID:           0,
		Status:           constants.OrderStatusCompleted,
		Currency:         "CNY",
		OriginalAmount:   models.NewMoneyFromDecimal(decimal.NewFromInt(40)),
		DiscountAmount:   models.NewMoneyFromDecimal(decimal.Zero),
		TotalAmount:      models.NewMoneyFromDecimal(decimal.NewFromInt(40)),
		WalletPaidAmount: models.NewMoneyFromDecimal(decimal.Zero),
		OnlinePaidAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(40)),
		RefundedAmount:   models.NewMoneyFromDecimal(decimal.Zero),
		PaidAt:           &paidAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.Create(childB).Error; err != nil {
		t.Fatalf("create childB order failed: %v", err)
	}

	svc := NewOrderService(OrderServiceOptions{
		OrderRepo: repository.NewOrderRepository(db),
	})
	updated, err := svc.UpdateOrderStatus(parent.ID, constants.OrderStatusPartiallyRefunded)
	if err != nil {
		t.Fatalf("update parent status failed: %v", err)
	}
	if updated == nil || updated.Status != constants.OrderStatusPartiallyRefunded {
		t.Fatalf("expected parent partially_refunded, got: %+v", updated)
	}
	if len(updated.Children) != 2 {
		t.Fatalf("expected 2 children in updated order, got: %d", len(updated.Children))
	}
	for _, child := range updated.Children {
		if child.Status != constants.OrderStatusPartiallyRefunded {
			t.Fatalf("expected child partially_refunded, got: %s", child.Status)
		}
	}

	var reloadedA models.Order
	if err := db.First(&reloadedA, childA.ID).Error; err != nil {
		t.Fatalf("reload childA failed: %v", err)
	}
	if reloadedA.Status != constants.OrderStatusPartiallyRefunded {
		t.Fatalf("expected childA partially_refunded, got: %s", reloadedA.Status)
	}
	var reloadedB models.Order
	if err := db.First(&reloadedB, childB.ID).Error; err != nil {
		t.Fatalf("reload childB failed: %v", err)
	}
	if reloadedB.Status != constants.OrderStatusPartiallyRefunded {
		t.Fatalf("expected childB partially_refunded, got: %s", reloadedB.Status)
	}
}

func TestCanCompleteParentOrder(t *testing.T) {
	order := &models.Order{
		Status: constants.OrderStatusDelivered,
		Children: []models.Order{
			{Status: constants.OrderStatusDelivered},
			{Status: constants.OrderStatusCompleted},
		},
	}
	if !canCompleteParentOrder(order) {
		t.Fatalf("expected delivered parent order to be completable")
	}
}

func TestCanCompleteParentOrderRejectInvalidStatus(t *testing.T) {
	order := &models.Order{
		Status: constants.OrderStatusPartiallyDelivered,
		Children: []models.Order{
			{Status: constants.OrderStatusDelivered},
		},
	}
	if canCompleteParentOrder(order) {
		t.Fatalf("expected partially_delivered parent order to be rejected")
	}
}

func TestCanCompleteParentOrderRejectInvalidChild(t *testing.T) {
	order := &models.Order{
		Status: constants.OrderStatusDelivered,
		Children: []models.Order{
			{Status: constants.OrderStatusPaid},
		},
	}
	if canCompleteParentOrder(order) {
		t.Fatalf("expected parent order with paid child to be rejected")
	}
}
