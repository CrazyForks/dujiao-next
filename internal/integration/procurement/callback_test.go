package procurement_test

import (
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/upstream"
)

type procurementCallbackStatusFixture struct {
	orderNo                   string
	initialOrderStatus        string
	initialProcurementStatus  string
	callbackStatus            string
	expectedProcurementStatus string
	expectedOrderStatus       string
}

func assertProcurementCallbackStatus(t *testing.T, fixture procurementCallbackStatusFixture) {
	t.Helper()
	db := setupProcurementTestDB(t)

	order := createProcTestOrder(t, db, fixture.orderNo, fixture.initialOrderStatus, constants.FulfillmentTypeUpstream)
	proc := createTestProcurementOrder(t, db, 1, order.ID, order.OrderNo, fixture.initialProcurementStatus)

	connSvc := NewSiteConnectionService(repository.NewSiteConnectionRepository(db), "test-key", t.TempDir())
	svc := newTestProcurementService(db, connSvc)

	if err := svc.HandleUpstreamCallback(proc.ID, fixture.callbackStatus, nil); err != nil {
		t.Fatalf("HandleUpstreamCallback: %v", err)
	}

	var updatedProc models.ProcurementOrder
	if err := db.First(&updatedProc, proc.ID).Error; err != nil {
		t.Fatalf("load procurement: %v", err)
	}
	if updatedProc.Status != fixture.expectedProcurementStatus {
		t.Errorf("expected procurement status %q, got %q", fixture.expectedProcurementStatus, updatedProc.Status)
	}

	var updatedOrder models.Order
	if err := db.First(&updatedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if updatedOrder.Status != fixture.expectedOrderStatus {
		t.Errorf("expected order status %q, got %q", fixture.expectedOrderStatus, updatedOrder.Status)
	}
}

// ── Phase 1 tests: order rollback on procurement failure ──

func TestRejectProcurement_RollsBackOrderStatus(t *testing.T) {
	db := setupProcurementTestDB(t)

	order := createProcTestOrder(t, db, "PROC-REJECT-001", constants.OrderStatusFulfilling, constants.FulfillmentTypeUpstream)
	proc := createTestProcurementOrder(t, db, 1, order.ID, order.OrderNo, "pending")

	connSvc := NewSiteConnectionService(repository.NewSiteConnectionRepository(db), "test-key", t.TempDir())
	svc := newTestProcurementService(db, connSvc)

	if err := svc.SubmitToUpstream(proc.ID); err != nil {
		t.Fatalf("SubmitToUpstream with missing connection: %v", err)
	}

	// 验证采购单状态 = rejected
	var updatedProc models.ProcurementOrder
	if err := db.First(&updatedProc, proc.ID).Error; err != nil {
		t.Fatalf("load procurement: %v", err)
	}
	if updatedProc.Status != "rejected" {
		t.Errorf("expected procurement status 'rejected', got %q", updatedProc.Status)
	}

	// 验证本地订单状态从 fulfilling 回退到 paid
	var updatedOrder models.Order
	if err := db.First(&updatedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if updatedOrder.Status != constants.OrderStatusPaid {
		t.Errorf("expected order status %q, got %q", constants.OrderStatusPaid, updatedOrder.Status)
	}
}

func TestHandleUpstreamCallback_Canceled_RollsBackOrder(t *testing.T) {
	db := setupProcurementTestDB(t)

	order := createProcTestOrder(t, db, "PROC-CANCEL-001", constants.OrderStatusFulfilling, constants.FulfillmentTypeUpstream)
	proc := createTestProcurementOrder(t, db, 1, order.ID, order.OrderNo, "accepted")

	connSvc := NewSiteConnectionService(repository.NewSiteConnectionRepository(db), "test-key", t.TempDir())
	svc := newTestProcurementService(db, connSvc)

	if err := svc.HandleUpstreamCallback(proc.ID, "canceled", nil); err != nil {
		t.Fatalf("HandleUpstreamCallback: %v", err)
	}

	// 验证采购单状态 = canceled
	var updatedProc models.ProcurementOrder
	if err := db.First(&updatedProc, proc.ID).Error; err != nil {
		t.Fatalf("load procurement: %v", err)
	}
	if updatedProc.Status != "canceled" {
		t.Errorf("expected procurement status 'canceled', got %q", updatedProc.Status)
	}

	// 验证本地订单状态从 fulfilling 回退到 paid
	var updatedOrder models.Order
	if err := db.First(&updatedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if updatedOrder.Status != constants.OrderStatusPaid {
		t.Errorf("expected order status %q, got %q", constants.OrderStatusPaid, updatedOrder.Status)
	}
}

func TestHandleUpstreamCallback_Delivered_CreatesFulfillment(t *testing.T) {
	db := setupProcurementTestDB(t)

	order := createProcTestOrder(t, db, "PROC-DELIVER-001", constants.OrderStatusFulfilling, constants.FulfillmentTypeUpstream)
	proc := createTestProcurementOrder(t, db, 1, order.ID, order.OrderNo, "accepted")

	connSvc := NewSiteConnectionService(repository.NewSiteConnectionRepository(db), "test-key", t.TempDir())
	svc := newTestProcurementService(db, connSvc)

	now := time.Now()
	fulfillment := &upstream.UpstreamFulfillment{
		Type:        constants.FulfillmentTypeUpstream,
		Status:      constants.FulfillmentStatusDelivered,
		Payload:     "CDK-001\nCDK-002",
		DeliveredAt: &now,
	}

	if err := svc.HandleUpstreamCallback(proc.ID, "delivered", fulfillment); err != nil {
		t.Fatalf("HandleUpstreamCallback: %v", err)
	}

	// 验证采购单状态 = fulfilled
	var updatedProc models.ProcurementOrder
	if err := db.First(&updatedProc, proc.ID).Error; err != nil {
		t.Fatalf("load procurement: %v", err)
	}
	if updatedProc.Status != "fulfilled" {
		t.Errorf("expected procurement status 'fulfilled', got %q", updatedProc.Status)
	}

	// 验证本地订单状态 = delivered
	var updatedOrder models.Order
	if err := db.First(&updatedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if updatedOrder.Status != constants.OrderStatusDelivered {
		t.Errorf("expected order status %q, got %q", constants.OrderStatusDelivered, updatedOrder.Status)
	}

	// 验证 Fulfillment 记录已创建
	var ff models.Fulfillment
	if err := db.Where("order_id = ?", order.ID).First(&ff).Error; err != nil {
		t.Fatalf("expected fulfillment record to exist: %v", err)
	}
	if ff.Payload != "CDK-001\nCDK-002" {
		t.Errorf("unexpected fulfillment payload: %q", ff.Payload)
	}
	if ff.Type != constants.FulfillmentTypeUpstream {
		t.Errorf("expected fulfillment type %q, got %q", constants.FulfillmentTypeUpstream, ff.Type)
	}
}

func TestHandleUpstreamCallback_PartiallyRefunded_AfterFulfilledUpdatesProcurementStatus(t *testing.T) {
	assertProcurementCallbackStatus(t, procurementCallbackStatusFixture{
		orderNo:                   "PROC-REFUND-KEEP-001",
		initialOrderStatus:        constants.OrderStatusDelivered,
		initialProcurementStatus:  constants.ProcurementStatusFulfilled,
		callbackStatus:            "partially_refunded",
		expectedProcurementStatus: constants.ProcurementStatusPartiallyRefunded,
		expectedOrderStatus:       constants.OrderStatusDelivered,
	})
}

func TestHandleUpstreamCallback_PartiallyRefunded_WhileFulfillingKeepsOrderStatus(t *testing.T) {
	assertProcurementCallbackStatus(t, procurementCallbackStatusFixture{
		orderNo:                   "PROC-REFUND-FULFILLING-001",
		initialOrderStatus:        constants.OrderStatusFulfilling,
		initialProcurementStatus:  constants.ProcurementStatusAccepted,
		callbackStatus:            "partially_refunded",
		expectedProcurementStatus: constants.ProcurementStatusPartiallyRefunded,
		expectedOrderStatus:       constants.OrderStatusFulfilling,
	})
}

func TestHandleUpstreamCallback_Refunded_AfterCompletedKeepsOrderStatus(t *testing.T) {
	assertProcurementCallbackStatus(t, procurementCallbackStatusFixture{
		orderNo:                   "PROC-REFUND-COMPLETED-001",
		initialOrderStatus:        constants.OrderStatusCompleted,
		initialProcurementStatus:  constants.ProcurementStatusFulfilled,
		callbackStatus:            "refunded",
		expectedProcurementStatus: constants.ProcurementStatusRefunded,
		expectedOrderStatus:       constants.OrderStatusCompleted,
	})
}
