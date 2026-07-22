package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/models"
	cataloggormstore "github.com/dujiao-next/internal/modules/catalog/store/gormstore"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/store/gormstore"
	"github.com/dujiao-next/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newAutoStockProductService(t *testing.T) (*ProductService, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:product_auto_stock_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.CardSecret{}); err != nil {
		t.Fatalf("auto migrate card secret failed: %v", err)
	}
	secretRepo := repository.NewCardSecretRepository(db)
	return NewProductService(nil, nil, secretRepo, nil, nil, nil, nil, nil, nil, nil), db
}

func insertCardSecrets(t *testing.T, db *gorm.DB, productID, skuID uint, status string, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		row := models.CardSecret{
			ProductID: productID,
			SKUID:     skuID,
			Secret:    fmt.Sprintf("secret-%d-%d-%s-%d", productID, skuID, status, i),
			Status:    status,
		}
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("create card secret failed: %v", err)
		}
	}
}

func newProductServiceForTest(t *testing.T) (*ProductService, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:product_service_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductSKU{}, &models.CardSecret{}, &models.CardSecretBatch{}, &models.MemberLevelPrice{}, &models.CartItem{}, &models.ProductMapping{}, &models.SKUMapping{}, &models.Order{}, &models.OrderItem{}, &models.PaymentChannel{}); err != nil {
		t.Fatalf("auto migrate product service tables failed: %v", err)
	}

	return NewProductService(
		repository.NewProductRepository(db),
		repository.NewProductSKURepository(db),
		repository.NewCardSecretRepository(db),
		repository.NewCardSecretBatchRepository(db),
		cataloggormstore.NewCategoryStore(db),
		memberlevelgormstore.NewPriceStore(db),
		repository.NewCartRepository(db),
		repository.NewProductMappingRepository(db),
		repository.NewOrderRepository(db),
		repository.NewPaymentChannelRepository(db),
	), db
}

func createProductTestPaymentChannel(t *testing.T, db *gorm.DB, name string, active bool, deleted bool) models.PaymentChannel {
	t.Helper()

	channel := models.PaymentChannel{
		Name:            name,
		ProviderType:    "official",
		ChannelType:     "wechat",
		InteractionMode: "qr",
		IsActive:        active,
	}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatalf("create payment channel failed: %v", err)
	}
	if !active {
		if err := db.Model(&channel).Update("is_active", false).Error; err != nil {
			t.Fatalf("disable payment channel failed: %v", err)
		}
		channel.IsActive = false
	}
	if deleted {
		if err := db.Delete(&channel).Error; err != nil {
			t.Fatalf("delete payment channel failed: %v", err)
		}
	}
	return channel
}
