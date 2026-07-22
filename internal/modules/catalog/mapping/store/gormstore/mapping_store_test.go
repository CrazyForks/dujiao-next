package gormstore

import (
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func setupMappingStoreTest(t *testing.T) (*MappingStore, *SKUMappingStore, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:mapping_store_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.ProductSKU{},
		&models.SiteConnection{},
		&models.ProductMapping{},
		&models.SKUMapping{},
	); err != nil {
		t.Fatalf("migrate mapping models failed: %v", err)
	}
	defaultCategory := models.Category{
		ID:       1,
		Slug:     "default-test-category",
		NameJSON: models.JSON{"zh-CN": "default"},
		IsActive: true,
	}
	if err := db.Create(&defaultCategory).Error; err != nil {
		t.Fatalf("seed default category failed: %v", err)
	}
	return NewMappingStore(db), NewSKUMappingStore(db), db
}

func createMappedProduct(t *testing.T, db *gorm.DB, slug, title string) *models.Product {
	t.Helper()
	product := &models.Product{
		CategoryID:      1,
		Slug:            slug,
		TitleJSON:       models.JSON{"zh-CN": title},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeUpstream,
		IsMapped:        true,
		IsActive:        true,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	return product
}

func createMapping(t *testing.T, store *MappingStore, connectionID, localProductID, upstreamProductID uint, upstreamStatus string, isActive bool) *models.ProductMapping {
	t.Helper()
	mapping := &models.ProductMapping{
		ConnectionID:      connectionID,
		LocalProductID:    localProductID,
		UpstreamProductID: upstreamProductID,
		UpstreamStatus:    upstreamStatus,
		IsActive:          true,
	}
	if err := store.Create(mapping); err != nil {
		t.Fatalf("create mapping failed: %v", err)
	}
	// IsActive 带 default:true 标签，Create 会忽略 false 零值，停用需走 Update
	if !isActive {
		mapping.IsActive = false
		if err := store.Update(mapping); err != nil {
			t.Fatalf("deactivate mapping failed: %v", err)
		}
	}
	return mapping
}

func TestMappingStoreListFiltersAndPaginates(t *testing.T) {
	store, _, db := setupMappingStoreTest(t)

	rechargeCard := createMappedProduct(t, db, "recharge-card", "上游充值卡")
	plainProduct := createMappedProduct(t, db, "plain-product", "普通商品")
	otherSite := createMappedProduct(t, db, "other-site", "另一站点商品")

	createMapping(t, store, 1, rechargeCard.ID, 1001, models.UpstreamStatusActive, true)
	inactive := createMapping(t, store, 1, plainProduct.ID, 1002, models.UpstreamStatusInactive, false)
	createMapping(t, store, 2, otherSite.ID, 2001, models.UpstreamStatusActive, true)

	if _, total, err := store.List(catalogmapping.ListFilter{}); err != nil || total != 3 {
		t.Fatalf("list all: total=%d err=%v, want 3", total, err)
	}
	if _, total, err := store.List(catalogmapping.ListFilter{ConnectionID: 1}); err != nil || total != 2 {
		t.Fatalf("list by connection: total=%d err=%v, want 2", total, err)
	}
	rows, total, err := store.List(catalogmapping.ListFilter{UpstreamStatus: models.UpstreamStatusInactive})
	if err != nil || total != 1 || len(rows) != 1 || rows[0].ID != inactive.ID {
		t.Fatalf("list by upstream status: rows=%d total=%d err=%v, want the inactive mapping", len(rows), total, err)
	}
	if _, total, err := store.List(catalogmapping.ListFilter{ProductStatus: "active"}); err != nil || total != 2 {
		t.Fatalf("list by product status active: total=%d err=%v, want 2", total, err)
	}
	if _, total, err := store.List(catalogmapping.ListFilter{ProductStatus: "inactive"}); err != nil || total != 1 {
		t.Fatalf("list by product status inactive: total=%d err=%v, want 1", total, err)
	}
	rows, total, err = store.List(catalogmapping.ListFilter{Search: "充值"})
	if err != nil || total != 1 || len(rows) != 1 || rows[0].LocalProductID != rechargeCard.ID {
		t.Fatalf("list by search: rows=%d total=%d err=%v, want the recharge card mapping", len(rows), total, err)
	}

	rows, total, err = store.List(catalogmapping.ListFilter{Page: 1, PageSize: 2})
	if err != nil || total != 3 || len(rows) != 2 {
		t.Fatalf("list page 1: rows=%d total=%d err=%v, want 2 rows of 3", len(rows), total, err)
	}
	rows, total, err = store.List(catalogmapping.ListFilter{Page: 2, PageSize: 2})
	if err != nil || total != 3 || len(rows) != 1 {
		t.Fatalf("list page 2: rows=%d total=%d err=%v, want 1 row of 3", len(rows), total, err)
	}
}

func TestMappingStoreDeleteByLocalProductRemovesSKUMappings(t *testing.T) {
	store, skuStore, db := setupMappingStoreTest(t)

	removed := createMappedProduct(t, db, "removed-product", "待删除商品")
	kept := createMappedProduct(t, db, "kept-product", "保留商品")
	removedMapping := createMapping(t, store, 1, removed.ID, 1001, models.UpstreamStatusActive, true)
	keptMapping := createMapping(t, store, 1, kept.ID, 1002, models.UpstreamStatusActive, true)

	for i, mappingID := range []uint{removedMapping.ID, removedMapping.ID, keptMapping.ID} {
		if err := skuStore.Create(&models.SKUMapping{
			ProductMappingID: mappingID,
			LocalSKUID:       uint(100 + i),
			UpstreamSKUID:    uint(200 + i),
		}); err != nil {
			t.Fatalf("create sku mapping failed: %v", err)
		}
	}

	if err := store.DeleteByLocalProduct(removed.ID); err != nil {
		t.Fatalf("delete by local product failed: %v", err)
	}

	if mapping, err := store.GetByLocalProductID(removed.ID); err != nil || mapping != nil {
		t.Fatalf("removed mapping should be gone: mapping=%v err=%v", mapping, err)
	}
	if rows, err := skuStore.ListByProductMapping(removedMapping.ID); err != nil || len(rows) != 0 {
		t.Fatalf("removed sku mappings should be gone: rows=%d err=%v", len(rows), err)
	}
	if mapping, err := store.GetByLocalProductID(kept.ID); err != nil || mapping == nil {
		t.Fatalf("kept mapping should survive: mapping=%v err=%v", mapping, err)
	}
	if rows, err := skuStore.ListByProductMapping(keptMapping.ID); err != nil || len(rows) != 1 {
		t.Fatalf("kept sku mapping should survive: rows=%d err=%v", len(rows), err)
	}
}
