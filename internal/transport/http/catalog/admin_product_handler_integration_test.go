package cataloghttp_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	cardsecretgormstore "github.com/dujiao-next/internal/modules/cardsecret/store/gormstore"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/store/gormstore"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	productapplication "github.com/dujiao-next/internal/modules/catalog/product/application"
	productadmin "github.com/dujiao-next/internal/modules/catalog/product/application/admin"
	productwrite "github.com/dujiao-next/internal/modules/catalog/product/application/write"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"
	cataloggormstore "github.com/dujiao-next/internal/modules/catalog/store/gormstore"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/store/gormstore"
	cataloghttp "github.com/dujiao-next/internal/transport/http/catalog"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type productHandlerFacade struct {
	*productapplication.Service
	*productadmin.AdminService
	*productwrite.WriteService
}

type productWriteUoW struct {
	products    *productgormstore.ProductStore
	skus        *productgormstore.SKUStore
	cardSecrets *cardsecretgormstore.Store
}

func (unit *productWriteUoW) WithinTransaction(fn func(productwrite.TransactionRepositories) error) error {
	if fn == nil {
		return nil
	}
	return unit.products.Transaction(func(tx *gorm.DB) error {
		return fn(productwrite.TransactionRepositories{
			Products:    unit.products.WithTx(tx),
			SKUs:        unit.skus.WithTx(tx),
			CardSecrets: unit.cardSecrets.WithTx(tx),
		})
	})
}

type productAdminUoW struct {
	products          *productgormstore.ProductStore
	productSKUs       *productgormstore.SKUStore
	cardSecrets       *cardsecretgormstore.Store
	cardSecretBatches *cardsecretgormstore.BatchStore
	memberLevelPrices *memberlevelgormstore.PriceStore
	carts             *cartDeleteStore
	productMappings   *mappinggormstore.MappingStore
}

type cartDeleteStore struct{ db *gorm.DB }

func (s *cartDeleteStore) DeleteByProduct(productID uint) error {
	return s.db.Where("product_id = ?", productID).Delete(&models.CartItem{}).Error
}

func (s *cartDeleteStore) WithTx(tx *gorm.DB) *cartDeleteStore {
	return &cartDeleteStore{db: tx}
}

func (unit *productAdminUoW) WithinTransaction(fn func(productadmin.DeleteRepositories) error) error {
	if fn == nil {
		return nil
	}
	return unit.products.Transaction(func(tx *gorm.DB) error {
		return fn(productadmin.DeleteRepositories{
			Products:          unit.products.WithTx(tx),
			CardSecrets:       unit.cardSecrets.WithTx(tx),
			CardSecretBatches: unit.cardSecretBatches.WithTx(tx),
			SKUs:              unit.productSKUs.WithTx(tx),
			MemberLevelPrices: memberlevelgormstore.NewPriceStore(tx),
			Carts:             unit.carts.WithTx(tx),
			ProductMappings:   unit.productMappings.WithTx(tx),
		})
	})
}

type orderHistoryStore struct{ db *gorm.DB }

func (s *orderHistoryStore) CountOrderItemsByProduct(productID uint) (int64, error) {
	var count int64
	err := s.db.Model(&models.OrderItem{}).Where("product_id = ?", productID).Count(&count).Error
	return count, err
}

type paymentChannelStore struct{ db *gorm.DB }

func (s *paymentChannelStore) ListByIDs(ids []uint) ([]models.PaymentChannel, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []models.PaymentChannel
	err := s.db.Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}

func setupAdminProductHandlerTest(t *testing.T) (*cataloghttp.AdminProductHandler, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:admin_product_handler_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.ProductSKU{},
		&models.CardSecret{},
		&models.CardSecretBatch{},
		&models.MemberLevelPrice{},
		&models.CartItem{},
		&models.ProductMapping{},
		&models.SKUMapping{},
		&models.Order{},
		&models.OrderItem{},
		&models.PaymentChannel{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	productStore := productgormstore.NewProductStore(db)
	skuStore := productgormstore.NewSKUStore(db)
	cardSecretStore := cardsecretgormstore.New(db)
	cardSecretBatchStore := cardsecretgormstore.NewBatch(db)
	categoryStore := cataloggormstore.NewCategoryStore(db)
	memberLevelPriceStore := memberlevelgormstore.NewPriceStore(db)
	mappingStore := mappinggormstore.NewMappingStore(db)
	skuMappingStore := mappinggormstore.NewSKUMappingStore(db)
	cartStore := &cartDeleteStore{db: db}

	facade := &productHandlerFacade{
		Service: productapplication.NewService(productapplication.Options{
			Products:                      productStore,
			Categories:                    categoryStore,
			Stock:                         cardSecretStore,
			NotFoundError:                 catalogproduct.ErrNotFound,
			ResellerProductNotListedError: catalogproduct.ErrResellerProductNotListed,
		}),
		AdminService: productadmin.NewAdminService(productadmin.Options{
			Products:    productStore,
			Categories:  categoryStore,
			CardSecrets: cardSecretStore,
			Orders:      &orderHistoryStore{db: db},
			Transactions: &productAdminUoW{
				products:          productStore,
				productSKUs:       skuStore,
				cardSecrets:       cardSecretStore,
				cardSecretBatches: cardSecretBatchStore,
				memberLevelPrices: memberLevelPriceStore,
				carts:             cartStore,
				productMappings:   mappingStore,
			},
			Errors: productadmin.ErrorSet{
				NotFound:               catalogproduct.ErrNotFound,
				ProductCategoryInvalid: catalogproduct.ErrProductCategoryInvalid,
				ProductHasStock:        catalogproduct.ErrProductHasStock,
				ProductHasOrderRecord:  catalogproduct.ErrProductHasOrderRecord,
			},
		}),
		WriteService: productwrite.NewWriteService(productwrite.Options{
			Products:        productStore,
			SKUs:            skuStore,
			Categories:      categoryStore,
			PaymentChannels: &paymentChannelStore{db: db},
			Transactions: &productWriteUoW{
				products:    productStore,
				skus:        skuStore,
				cardSecrets: cardSecretStore,
			},
			Errors: productwrite.ErrorSet{
				NotFound:                     catalogproduct.ErrNotFound,
				SlugExists:                   catalogproduct.ErrSlugExists,
				ProductCategoryInvalid:       catalogproduct.ErrProductCategoryInvalid,
				ProductPurchaseInvalid:       catalogproduct.ErrProductPurchaseInvalid,
				FulfillmentInvalid:           catalogproduct.ErrFulfillmentInvalid,
				ProductPriceInvalid:          catalogproduct.ErrProductPriceInvalid,
				ManualStockInvalid:           catalogproduct.ErrManualStockInvalid,
				ProductPurchaseLimitInvalid:  catalogproduct.ErrProductPurchaseLimitInvalid,
				ProductStockDisplayInvalid:   catalogproduct.ErrProductStockDisplayInvalid,
				ProductSKUInvalid:            catalogproduct.ErrProductSKUInvalid,
				ProductSKUHasCardSecretStock: catalogproduct.ErrProductSKUHasCardSecretStock,
			},
		}),
	}

	h := cataloghttp.NewAdminProductHandler(facade, facade, nil, mappingStore, skuMappingStore)
	return h, db
}

func TestBatchUpdateProductStatusReturnsFailureReasons(t *testing.T) {
	h, db := setupAdminProductHandlerTest(t)

	product := models.Product{
		CategoryID:      0,
		Slug:            "batch-uncategorized-product",
		TitleJSON:       models.JSON{"zh-CN": "batch-uncategorized-product"},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		FulfillmentType: constants.FulfillmentTypeUpstream,
		IsMapped:        true,
		IsActive:        false,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create uncategorized product failed: %v", err)
	}

	body := fmt.Sprintf(`{"ids":[%d],"is_active":true}`, product.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/products/batch-status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.BatchUpdateProductStatus(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Total        int `json:"total"`
			SuccessCount int `json:"success_count"`
			FailedItems  []struct {
				ID        uint   `json:"id"`
				ErrorCode string `json:"error_code"`
				Message   string `json:"message"`
			} `json:"failed_items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, w.Body.String())
	}
	if resp.Data.Total != 1 || resp.Data.SuccessCount != 0 {
		t.Fatalf("unexpected counts: total=%d success=%d", resp.Data.Total, resp.Data.SuccessCount)
	}
	if len(resp.Data.FailedItems) != 1 {
		t.Fatalf("expected one failed item, got %+v", resp.Data.FailedItems)
	}
	if resp.Data.FailedItems[0].ID != product.ID {
		t.Fatalf("expected failed product id %d, got %d", product.ID, resp.Data.FailedItems[0].ID)
	}
	if resp.Data.FailedItems[0].ErrorCode != "product_category_invalid" {
		t.Fatalf("expected product_category_invalid, got %q", resp.Data.FailedItems[0].ErrorCode)
	}
}

func TestUpdateProductWholesalePricesHandlerUpdatesTiers(t *testing.T) {
	h, db := setupAdminProductHandlerTest(t)

	product := models.Product{
		CategoryID:  1,
		Slug:        "handler-wholesale-product",
		TitleJSON:   models.JSON{"zh-CN": "handler-wholesale-product"},
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		IsActive:    true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	body := `{"wholesale_prices":[{"min_quantity":10,"unit_price":70},{"min_quantity":5,"unit_price":80}]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/products/1/wholesale-prices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", product.ID)}}

	h.UpdateProductWholesalePrices(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data models.Product `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, w.Body.String())
	}
	if len(resp.Data.WholesalePrices) != 2 {
		t.Fatalf("expected 2 wholesale tiers, got %+v", resp.Data.WholesalePrices)
	}
	if resp.Data.WholesalePrices[0].MinQuantity != 5 || resp.Data.WholesalePrices[0].UnitPrice.String() != "80.00" {
		t.Fatalf("expected sorted first tier min=5 price=80.00, got %+v", resp.Data.WholesalePrices[0])
	}
}

func TestUpdateProductWholesalePricesHandlerAllowsClear(t *testing.T) {
	h, db := setupAdminProductHandlerTest(t)

	product := models.Product{
		CategoryID:  1,
		Slug:        "handler-wholesale-clear",
		TitleJSON:   models.JSON{"zh-CN": "handler-wholesale-clear"},
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		WholesalePrices: models.WholesalePriceTiers{
			{MinQuantity: 5, UnitPrice: models.NewMoneyFromDecimal(decimal.NewFromInt(80))},
		},
		IsActive: true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/products/1/wholesale-prices", strings.NewReader(`{"wholesale_prices":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", product.ID)}}

	h.UpdateProductWholesalePrices(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	var got models.Product
	if err := db.First(&got, product.ID).Error; err != nil {
		t.Fatalf("reload product failed: %v", err)
	}
	if len(got.WholesalePrices) != 0 {
		t.Fatalf("expected wholesale prices cleared, got %+v", got.WholesalePrices)
	}
}

func TestUpdateProductWholesalePricesHandlerRejectsInvalidTier(t *testing.T) {
	h, db := setupAdminProductHandlerTest(t)

	product := models.Product{
		CategoryID:  1,
		Slug:        "handler-wholesale-invalid",
		TitleJSON:   models.JSON{"zh-CN": "handler-wholesale-invalid"},
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		IsActive:    true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/products/1/wholesale-prices", strings.NewReader(`{"wholesale_prices":[{"min_quantity":0,"unit_price":80}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", product.ID)}}

	h.UpdateProductWholesalePrices(c)

	if w.Code != http.StatusOK {
		t.Fatalf("project response wrapper should still return HTTP 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		StatusCode int    `json:"status_code"`
		Msg        string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, w.Body.String())
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected wholesale invalid response, got status_code=%d body=%s", resp.StatusCode, w.Body.String())
	}
}

func TestUpdateProductWholesalePricesHandlerReturnsNotFound(t *testing.T) {
	h, _ := setupAdminProductHandlerTest(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/products/999999/wholesale-prices", strings.NewReader(`{"wholesale_prices":[{"min_quantity":5,"unit_price":80}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999999"}}

	h.UpdateProductWholesalePrices(c)

	if w.Code != http.StatusOK {
		t.Fatalf("project response wrapper should still return HTTP 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		StatusCode int `json:"status_code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, w.Body.String())
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected product not found response, got body=%s", w.Body.String())
	}
}
