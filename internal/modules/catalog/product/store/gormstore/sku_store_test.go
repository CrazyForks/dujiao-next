package gormstore

import (
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func setupSKUStoreTest(t *testing.T) (*SKUStore, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:product_sku_repository_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.ProductSKU{}); err != nil {
		t.Fatalf("migrate product sku failed: %v", err)
	}
	return NewSKUStore(db), db
}

func TestSKUStoreListByProductSortOrderDescending(t *testing.T) {
	repo, _ := setupSKUStoreTest(t)

	high := &models.ProductSKU{
		ProductID:      1,
		SKUCode:        "HIGH",
		PriceAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		IsActive:       true,
		SortOrder:      100,
		SpecValuesJSON: jsonmap.JSON{},
	}
	low := &models.ProductSKU{
		ProductID:      1,
		SKUCode:        "LOW",
		PriceAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		IsActive:       true,
		SortOrder:      1,
		SpecValuesJSON: jsonmap.JSON{},
	}
	if err := repo.Create(high); err != nil {
		t.Fatalf("create high sort sku failed: %v", err)
	}
	if err := repo.Create(low); err != nil {
		t.Fatalf("create low sort sku failed: %v", err)
	}

	rows, err := repo.ListByProduct(1, true)
	if err != nil {
		t.Fatalf("list skus failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 skus, got %d", len(rows))
	}
	if rows[0].SKUCode != "HIGH" || rows[1].SKUCode != "LOW" {
		t.Fatalf("expected high sort_order first, got %s then %s", rows[0].SKUCode, rows[1].SKUCode)
	}
}

func TestProductSKUManualStockLifecycleMatchesProductSemantics(t *testing.T) {
	repo, db := setupSKUStoreTest(t)
	sku := &models.ProductSKU{
		ProductID:        1,
		SKUCode:          "STOCK-LIFECYCLE",
		PriceAmount:      models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		ManualStockTotal: 10,
		IsActive:         true,
	}
	if err := repo.Create(sku); err != nil {
		t.Fatalf("create stock sku: %v", err)
	}

	assertAffected := func(operation string, affected int64, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s manual stock: %v", operation, err)
		}
		if affected != 1 {
			t.Fatalf("%s affected rows want 1 got %d", operation, affected)
		}
	}

	affected, err := repo.ReserveManualStock(sku.ID, 3)
	assertAffected("reserve", affected, err)
	affected, err = repo.ConsumeManualStock(sku.ID, 2)
	assertAffected("consume", affected, err)
	affected, err = repo.ReleaseManualStock(sku.ID, 1)
	assertAffected("release", affected, err)

	var reloaded models.ProductSKU
	if err := db.First(&reloaded, sku.ID).Error; err != nil {
		t.Fatalf("reload stock sku: %v", err)
	}
	if reloaded.ManualStockTotal != 8 || reloaded.ManualStockLocked != 0 || reloaded.ManualStockSold != 2 {
		t.Fatalf("unexpected stock lifecycle result: %#v", reloaded)
	}

	unlimited := &models.ProductSKU{
		ProductID:        1,
		SKUCode:          "STOCK-UNLIMITED",
		PriceAmount:      models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		ManualStockTotal: constants.ManualStockUnlimited,
		IsActive:         true,
	}
	if err := repo.Create(unlimited); err != nil {
		t.Fatalf("create unlimited sku: %v", err)
	}
	if affected, err := repo.ReserveManualStock(unlimited.ID, 1); err != nil || affected != 0 {
		t.Fatalf("unlimited reserve should be no-op, affected=%d err=%v", affected, err)
	}
	if affected, err := repo.ConsumeManualStock(unlimited.ID, 1); err != nil || affected != 0 {
		t.Fatalf("unlimited consume should be no-op, affected=%d err=%v", affected, err)
	}
}
