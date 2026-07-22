package integrationtest

import (
	"fmt"
	"testing"
	"time"

	memberleveldomain "github.com/dujiao-next/internal/modules/memberlevel/domain"

	mappingdomain "github.com/dujiao-next/internal/modules/catalog/mapping/domain"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"

	categorydomain "github.com/dujiao-next/internal/modules/catalog/category/domain"

	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"

	catalogproductbootstrap "github.com/dujiao-next/internal/bootstrap/catalogproduct"
	"github.com/dujiao-next/internal/models"
	cartdomain "github.com/dujiao-next/internal/modules/cart/domain"
	cartgormstore "github.com/dujiao-next/internal/modules/cart/infrastructure/gormstore"
	categorygormstore "github.com/dujiao-next/internal/modules/catalog/category/infrastructure/gormstore"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/infrastructure/gormstore"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/infrastructure/gormstore"
	"github.com/dujiao-next/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newAutoStockProductService(t *testing.T) (catalogproductbootstrap.Services, *gorm.DB) {
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
	return catalogproductbootstrap.New(catalogproductbootstrap.Dependencies{CardSecrets: secretRepo}), db
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

func newProductServiceForTest(t *testing.T) (catalogproductbootstrap.Services, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:product_service_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&categorydomain.Category{}, &productdomain.Product{}, &productdomain.ProductSKU{}, &models.CardSecret{}, &models.CardSecretBatch{}, &memberleveldomain.MemberLevelPrice{}, &cartdomain.Item{}, &mappingdomain.Mapping{}, &mappingdomain.SKUMapping{}, &models.Order{}, &models.OrderItem{}, &models.PaymentChannel{}); err != nil {
		t.Fatalf("auto migrate product service tables failed: %v", err)
	}

	return catalogproductbootstrap.New(catalogproductbootstrap.Dependencies{
		Products:          productgormstore.NewProductStore(db),
		SKUs:              productgormstore.NewSKUStore(db),
		CardSecrets:       repository.NewCardSecretRepository(db),
		CardSecretBatches: repository.NewCardSecretBatchRepository(db),
		Categories:        categorygormstore.NewCategoryStore(db),
		MemberLevelPrices: memberlevelgormstore.NewPriceStore(db),
		Carts:             cartgormstore.New(db),
		ProductMappings:   mappinggormstore.NewMappingStore(db),
		Orders:            repository.NewOrderRepository(db),
		PaymentChannels:   repository.NewPaymentChannelRepository(db),
	}), db
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
