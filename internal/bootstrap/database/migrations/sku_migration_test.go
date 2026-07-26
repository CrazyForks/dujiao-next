package migrations

import (
	"fmt"
	"testing"
	"time"

	cardsecretdomain "github.com/dujiao-next/internal/modules/cardsecret/domain"
	cartdomain "github.com/dujiao-next/internal/modules/cart/domain"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	settingsstore "github.com/dujiao-next/internal/modules/settings/infrastructure/gormstore"
	"github.com/dujiao-next/internal/platform/database/gormdb"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func setupSKUMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:sku_migration_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	gormdb.DB = db
	return db
}

func TestEnsureProductSKUMigrationBackfillLegacyData(t *testing.T) {
	db := setupSKUMigrationTestDB(t)

	if err := db.AutoMigrate(
		&productdomain.Product{},
		&productdomain.ProductSKU{},
		&orderdomain.OrderItem{},
		&cartdomain.Item{},
		&cardsecretdomain.Secret{},
		&cardsecretdomain.Batch{},
		&settingsstore.SettingRecord{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	product := &productdomain.Product{
		CategoryID:        1,
		Slug:              "sku-migration-legacy",
		TitleJSON:         jsonmap.JSON{"zh-CN": "历史商品"},
		PriceAmount:       money.FromDecimal(decimal.NewFromInt(128)),
		PurchaseType:      "member",
		FulfillmentType:   "manual",
		ManualStockTotal:  20,
		ManualStockLocked: 3,
		ManualStockSold:   5,
		IsActive:          true,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	now := time.Now()
	orderItem := &orderdomain.OrderItem{
		OrderID:         1,
		ProductID:       product.ID,
		SKUID:           0,
		TitleJSON:       jsonmap.JSON{"zh-CN": "历史商品"},
		UnitPrice:       product.PriceAmount,
		Quantity:        1,
		TotalPrice:      product.PriceAmount,
		FulfillmentType: "manual",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(orderItem).Error; err != nil {
		t.Fatalf("create order item failed: %v", err)
	}

	cartItem := &cartdomain.Item{
		UserID:          1001,
		ProductID:       product.ID,
		SKUID:           0,
		Quantity:        2,
		FulfillmentType: "manual",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(cartItem).Error; err != nil {
		t.Fatalf("create cart item failed: %v", err)
	}

	batch := &cardsecretdomain.Batch{
		ProductID:  product.ID,
		SKUID:      0,
		BatchNo:    "SKU-MIGRATION-BATCH-001",
		Source:     "manual",
		TotalCount: 1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := db.Create(batch).Error; err != nil {
		t.Fatalf("create card secret batch failed: %v", err)
	}

	secret := &cardsecretdomain.Secret{
		ProductID: product.ID,
		SKUID:     0,
		BatchID:   &batch.ID,
		Secret:    "CARD-001",
		Status:    cardsecretdomain.StatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(secret).Error; err != nil {
		t.Fatalf("create card secret failed: %v", err)
	}

	if err := ensureProductSKUMigration(); err != nil {
		t.Fatalf("ensure sku migration failed: %v", err)
	}

	var sku productdomain.ProductSKU
	if err := db.Where("product_id = ? AND sku_code = ?", product.ID, productdomain.DefaultSKUCode).First(&sku).Error; err != nil {
		t.Fatalf("query default sku failed: %v", err)
	}
	if !sku.PriceAmount.Decimal.Equal(product.PriceAmount.Decimal) {
		t.Fatalf("default sku price mismatch want %s got %s", product.PriceAmount.String(), sku.PriceAmount.String())
	}
	if sku.ManualStockTotal != product.ManualStockTotal || sku.ManualStockLocked != product.ManualStockLocked || sku.ManualStockSold != product.ManualStockSold {
		t.Fatalf("default sku stock snapshot mismatch")
	}

	var gotOrderItem orderdomain.OrderItem
	if err := db.First(&gotOrderItem, orderItem.ID).Error; err != nil {
		t.Fatalf("reload order item failed: %v", err)
	}
	if gotOrderItem.SKUID != sku.ID {
		t.Fatalf("order item sku_id want %d got %d", sku.ID, gotOrderItem.SKUID)
	}

	var gotCartItem cartdomain.Item
	if err := db.First(&gotCartItem, cartItem.ID).Error; err != nil {
		t.Fatalf("reload cart item failed: %v", err)
	}
	if gotCartItem.SKUID != sku.ID {
		t.Fatalf("cart item sku_id want %d got %d", sku.ID, gotCartItem.SKUID)
	}

	var gotBatch cardsecretdomain.Batch
	if err := db.First(&gotBatch, batch.ID).Error; err != nil {
		t.Fatalf("reload card secret batch failed: %v", err)
	}
	if gotBatch.SKUID != sku.ID {
		t.Fatalf("card secret batch sku_id want %d got %d", sku.ID, gotBatch.SKUID)
	}

	var gotSecret cardsecretdomain.Secret
	if err := db.First(&gotSecret, secret.ID).Error; err != nil {
		t.Fatalf("reload card secret failed: %v", err)
	}
	if gotSecret.SKUID != sku.ID {
		t.Fatalf("card secret sku_id want %d got %d", sku.ID, gotSecret.SKUID)
	}

	if err := ensureProductSKUMigration(); err != nil {
		t.Fatalf("ensure sku migration second run failed: %v", err)
	}

	var skuCount int64
	if err := db.Model(&productdomain.ProductSKU{}).Where("product_id = ?", product.ID).Count(&skuCount).Error; err != nil {
		t.Fatalf("count product sku failed: %v", err)
	}
	if skuCount != 1 {
		t.Fatalf("idempotent check failed: sku count want 1 got %d", skuCount)
	}
}

func TestMigrateCartSKUUniqueIndex(t *testing.T) {
	db := setupSKUMigrationTestDB(t)
	if err := db.AutoMigrate(&cartdomain.Item{}); err != nil {
		t.Fatalf("auto migrate cart item failed: %v", err)
	}

	// 构造历史唯一索引，验证迁移函数会移除该索引并保留新索引。
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_user_product ON cart_items(user_id, product_id)").Error; err != nil {
		t.Fatalf("create legacy index failed: %v", err)
	}

	if err := migrateCartSKUUniqueIndex(); err != nil {
		t.Fatalf("migrate cart unique index failed: %v", err)
	}

	if db.Migrator().HasIndex(&cartdomain.Item{}, "idx_cart_user_product") {
		t.Fatalf("legacy unique index idx_cart_user_product should be dropped")
	}
	if !db.Migrator().HasIndex(&cartdomain.Item{}, "idx_cart_user_product_sku") {
		t.Fatalf("new unique index idx_cart_user_product_sku should exist")
	}
}
