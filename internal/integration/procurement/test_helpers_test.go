package procurement_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/procurement"
	procurementgormstore "github.com/dujiao-next/internal/modules/procurement/store/gormstore"
	"github.com/dujiao-next/internal/modules/siteconnection"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type SiteConnectionService = siteconnection.Service
type CreateConnectionInput = siteconnection.CreateInput
type ListFilter = procurement.ListFilter

var (
	NewSiteConnectionService = siteconnection.NewService
	ErrExists                = procurement.ErrExists
)

func setupProcurementTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:procurement_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Order{},
		&models.OrderItem{},
		&models.OrderRefundRecord{},
		&models.Fulfillment{},
		&models.ProcurementOrder{},
		&models.SiteConnection{},
		&models.ProductMapping{},
		&models.SKUMapping{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	models.DB = db
	return db
}

func createProcTestOrder(t *testing.T, db *gorm.DB, orderNo, status, fulfillmentType string) *models.Order {
	t.Helper()
	order := &models.Order{
		OrderNo:        orderNo,
		UserID:         1,
		Status:         status,
		Currency:       "CNY",
		OriginalAmount: money.FromDecimal(decimal.NewFromInt(100)),
		TotalAmount:    money.FromDecimal(decimal.NewFromInt(100)),
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	item := &models.OrderItem{
		OrderID:         order.ID,
		ProductID:       1,
		SKUID:           1,
		Quantity:        1,
		FulfillmentType: fulfillmentType,
		TitleJSON:       jsonmap.JSON{"zh-CN": "Test Product"},
		UnitPrice:       money.FromDecimal(decimal.NewFromInt(100)),
		TotalPrice:      money.FromDecimal(decimal.NewFromInt(100)),
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("create order item failed: %v", err)
	}
	var loaded models.Order
	if err := db.Preload("Items").First(&loaded, order.ID).Error; err != nil {
		t.Fatalf("reload order failed: %v", err)
	}
	return &loaded
}

func createTestProcurementOrder(t *testing.T, db *gorm.DB, connID, localOrderID uint, localOrderNo, status string) *models.ProcurementOrder {
	t.Helper()
	order := &models.ProcurementOrder{
		ConnectionID:    connID,
		LocalOrderID:    localOrderID,
		LocalOrderNo:    localOrderNo,
		Status:          status,
		LocalSellAmount: money.FromDecimal(decimal.NewFromInt(100)),
		Currency:        "CNY",
		TraceID:         "test-trace-id",
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create procurement order failed: %v", err)
	}
	return order
}

func newTestProcurementService(db *gorm.DB, connections *SiteConnectionService) *procurement.Service {
	orders := repository.NewOrderRepository(db)
	fulfillments := repository.NewFulfillmentRepository(db)
	return procurement.NewService(procurement.ServiceOptions{
		Repository:      procurementgormstore.New(db),
		Orders:          orders,
		ProductMappings: repository.NewProductMappingRepository(db),
		SKUMappings:     repository.NewSKUMappingRepository(db),
		Connections:     connections,
		OrderLifecycle: service.NewProcurementOrderLifecycle(
			orders,
			fulfillments,
			nil,
			nil,
			config.EmailConfig{},
		),
	})
}
