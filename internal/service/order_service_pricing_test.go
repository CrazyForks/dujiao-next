package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	coupongormstore "github.com/dujiao-next/internal/modules/coupon/store/gormstore"
	"github.com/dujiao-next/internal/modules/memberlevel"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/store/gormstore"
	promotiongormstore "github.com/dujiao-next/internal/modules/promotion/store/gormstore"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/shared/jsonmap"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type orderPurchaseQuantityLimitFixture struct {
	dsnPrefix       string
	categorySlug    string
	productSlug     string
	minQuantity     int
	maxQuantity     int
	requestQuantity int
	expectedErr     error
}

func assertBuildOrderResultRejectsPurchaseQuantity(t *testing.T, fixture orderPurchaseQuantityLimitFixture) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", fixture.dsnPrefix, time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductSKU{}, &models.Promotion{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now()
	category := models.Category{
		Slug:      fixture.categorySlug,
		NameJSON:  jsonmap.JSON{"zh-CN": "测试分类"},
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	product := models.Product{
		CategoryID:          category.ID,
		Slug:                fixture.productSlug,
		TitleJSON:           jsonmap.JSON{"zh-CN": "测试商品"},
		PriceAmount:         models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		PurchaseType:        constants.ProductPurchaseMember,
		FulfillmentType:     constants.FulfillmentTypeManual,
		MinPurchaseQuantity: fixture.minQuantity,
		MaxPurchaseQuantity: fixture.maxQuantity,
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	sku := models.ProductSKU{
		ProductID:         product.ID,
		SKUCode:           models.DefaultSKUCode,
		PriceAmount:       models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		IsActive:          true,
		ManualStockTotal:  constants.ManualStockUnlimited,
		ManualStockLocked: 0,
		ManualStockSold:   0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	svc := NewOrderService(OrderServiceOptions{
		ProductRepo:    repository.NewProductRepository(db),
		ProductSKURepo: repository.NewProductSKURepository(db),
		PromotionRepo:  promotiongormstore.New(db),
		ExpireMinutes:  15,
	})

	_, err = svc.buildOrderResult(orderCreateParams{
		UserID: 1,
		Items: []CreateOrderItem{
			{
				ProductID: product.ID,
				SKUID:     sku.ID,
				Quantity:  fixture.requestQuantity,
			},
		},
	})
	if !errors.Is(err, fixture.expectedErr) {
		t.Fatalf("expected %v, got: %v", fixture.expectedErr, err)
	}
}

func TestBuildOrderResultRejectsZeroPromotionPrice(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_promo_zero_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductSKU{}, &models.Promotion{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now()
	category := models.Category{
		Slug:      "test-category",
		NameJSON:  jsonmap.JSON{"zh-CN": "测试分类"},
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	product := models.Product{
		CategoryID:      category.ID,
		Slug:            "test-product",
		TitleJSON:       jsonmap.JSON{"zh-CN": "测试商品"},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeManual,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	sku := models.ProductSKU{
		ProductID:         product.ID,
		SKUCode:           models.DefaultSKUCode,
		PriceAmount:       models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		IsActive:          true,
		ManualStockTotal:  constants.ManualStockUnlimited,
		ManualStockLocked: 0,
		ManualStockSold:   0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	promotion := models.Promotion{
		Name:       "test-100-percent",
		ScopeType:  constants.ScopeTypeProduct,
		ScopeRefID: product.ID,
		Type:       constants.PromotionTypePercent,
		Value:      models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		MinAmount:  models.NewMoneyFromDecimal(decimal.Zero),
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := db.Create(&promotion).Error; err != nil {
		t.Fatalf("create promotion failed: %v", err)
	}

	svc := NewOrderService(OrderServiceOptions{
		ProductRepo:    repository.NewProductRepository(db),
		ProductSKURepo: repository.NewProductSKURepository(db),
		PromotionRepo:  promotiongormstore.New(db),
		ExpireMinutes:  15,
	})

	_, err = svc.buildOrderResult(orderCreateParams{
		UserID: 1,
		Items: []CreateOrderItem{
			{
				ProductID: product.ID,
				SKUID:     sku.ID,
				Quantity:  1,
			},
		},
	})
	if !errors.Is(err, ErrProductPriceInvalid) {
		t.Fatalf("expected product price invalid, got: %v", err)
	}
}

func TestPreviewOrderAppliesMemberDiscountForManualProductBeforeFormCompleted(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_manual_member_preview_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.ProductSKU{},
		&models.Promotion{},
		&models.User{},
		&models.MemberLevel{},
		&models.MemberLevelPrice{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now()
	category := models.Category{
		Slug:      "manual-member-preview-category",
		NameJSON:  jsonmap.JSON{"zh-CN": "测试分类"},
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	level := models.MemberLevel{
		NameJSON:     jsonmap.JSON{"zh-CN": "金牌会员"},
		Slug:         "gold",
		DiscountRate: models.NewMoneyFromDecimal(decimal.NewFromInt(80)),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&level).Error; err != nil {
		t.Fatalf("create member level failed: %v", err)
	}
	user := models.User{
		Email:         "manual-preview@example.com",
		PasswordHash:  "hash",
		Status:        "active",
		MemberLevelID: level.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	product := models.Product{
		CategoryID:      category.ID,
		Slug:            "manual-member-preview-product",
		TitleJSON:       jsonmap.JSON{"zh-CN": "人工发货商品"},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeManual,
		ManualFormSchemaJSON: jsonmap.JSON{
			"fields": []interface{}{
				map[string]interface{}{
					"key":      "account",
					"type":     "text",
					"required": true,
					"label":    map[string]interface{}{"zh-CN": "账号"},
				},
			},
		},
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	sku := models.ProductSKU{
		ProductID:         product.ID,
		SKUCode:           models.DefaultSKUCode,
		PriceAmount:       models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		IsActive:          true,
		ManualStockTotal:  constants.ManualStockUnlimited,
		ManualStockLocked: 0,
		ManualStockSold:   0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	levelRepo := memberlevelgormstore.NewLevelStore(db)
	priceRepo := memberlevelgormstore.NewPriceStore(db)
	userRepo := repository.NewUserRepository(db)
	svc := NewOrderService(OrderServiceOptions{
		UserRepo:           userRepo,
		ProductRepo:        repository.NewProductRepository(db),
		ProductSKURepo:     repository.NewProductSKURepository(db),
		PromotionRepo:      promotiongormstore.New(db),
		MemberLevelService: memberlevel.NewService(levelRepo, priceRepo, userRepo),
		ExpireMinutes:      15,
	})

	preview, err := svc.PreviewOrder(CreateOrderInput{
		UserID: user.ID,
		Items: []CreateOrderItem{
			{
				ProductID: product.ID,
				SKUID:     sku.ID,
				Quantity:  2,
			},
		},
	})
	if err != nil {
		t.Fatalf("preview order failed: %v", err)
	}

	expectedOriginal := decimal.NewFromInt(200)
	expectedMemberDiscount := decimal.NewFromInt(40)
	expectedTotal := decimal.NewFromInt(160)
	if !preview.OriginalAmount.Decimal.Equal(expectedOriginal) {
		t.Fatalf("expected original amount %s, got: %s", expectedOriginal.String(), preview.OriginalAmount.String())
	}
	if !preview.MemberDiscountAmount.Decimal.Equal(expectedMemberDiscount) {
		t.Fatalf("expected member discount amount %s, got: %s", expectedMemberDiscount.String(), preview.MemberDiscountAmount.String())
	}
	if !preview.TotalAmount.Decimal.Equal(expectedTotal) {
		t.Fatalf("expected total amount %s, got: %s", expectedTotal.String(), preview.TotalAmount.String())
	}
}

func TestBuildOrderResultStacksPromotionAndMemberDiscount(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_stack_promo_member_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.ProductSKU{},
		&models.Promotion{},
		&models.User{},
		&models.MemberLevel{},
		&models.MemberLevelPrice{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now()
	category := models.Category{
		Slug:      "stack-promo-member-category",
		NameJSON:  jsonmap.JSON{"zh-CN": "测试分类"},
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	level := models.MemberLevel{
		NameJSON:     jsonmap.JSON{"zh-CN": "金牌会员"},
		Slug:         "stack-gold",
		DiscountRate: models.NewMoneyFromDecimal(decimal.NewFromInt(80)),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&level).Error; err != nil {
		t.Fatalf("create member level failed: %v", err)
	}
	user := models.User{
		Email:         "stack-promo-member@example.com",
		PasswordHash:  "hash",
		Status:        "active",
		MemberLevelID: level.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	product := models.Product{
		CategoryID:      category.ID,
		Slug:            "stack-promo-member-product",
		TitleJSON:       jsonmap.JSON{"zh-CN": "叠加优惠商品"},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeAuto,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	sku := models.ProductSKU{
		ProductID:   product.ID,
		SKUCode:     models.DefaultSKUCode,
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}
	promotion := models.Promotion{
		Name:       "test-10-percent",
		ScopeType:  constants.ScopeTypeProduct,
		ScopeRefID: product.ID,
		Type:       constants.PromotionTypePercent,
		Value:      models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		MinAmount:  models.NewMoneyFromDecimal(decimal.Zero),
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := db.Create(&promotion).Error; err != nil {
		t.Fatalf("create promotion failed: %v", err)
	}

	levelRepo := memberlevelgormstore.NewLevelStore(db)
	priceRepo := memberlevelgormstore.NewPriceStore(db)
	userRepo := repository.NewUserRepository(db)
	svc := NewOrderService(OrderServiceOptions{
		UserRepo:           userRepo,
		ProductRepo:        repository.NewProductRepository(db),
		ProductSKURepo:     repository.NewProductSKURepository(db),
		PromotionRepo:      promotiongormstore.New(db),
		MemberLevelService: memberlevel.NewService(levelRepo, priceRepo, userRepo),
		ExpireMinutes:      15,
	})

	result, err := svc.buildOrderResult(orderCreateParams{
		UserID: user.ID,
		Items: []CreateOrderItem{
			{
				ProductID: product.ID,
				SKUID:     sku.ID,
				Quantity:  2,
			},
		},
	})
	if err != nil {
		t.Fatalf("buildOrderResult failed: %v", err)
	}

	expectedOriginal := decimal.NewFromInt(200)
	expectedPromotion := decimal.NewFromInt(20)
	expectedMemberDiscount := decimal.NewFromInt(36)
	expectedTotal := decimal.NewFromInt(144)
	if !result.OriginalAmount.Equal(expectedOriginal) {
		t.Fatalf("expected original amount %s, got: %s", expectedOriginal.String(), result.OriginalAmount.String())
	}
	if !result.PromotionDiscountAmount.Equal(expectedPromotion) {
		t.Fatalf("expected promotion discount amount %s, got: %s", expectedPromotion.String(), result.PromotionDiscountAmount.String())
	}
	if !result.MemberDiscountAmount.Equal(expectedMemberDiscount) {
		t.Fatalf("expected member discount amount %s, got: %s", expectedMemberDiscount.String(), result.MemberDiscountAmount.String())
	}
	if !result.TotalAmount.Equal(expectedTotal) {
		t.Fatalf("expected total amount %s, got: %s", expectedTotal.String(), result.TotalAmount.String())
	}
	if len(result.Plans) != 1 {
		t.Fatalf("expected one plan, got %d", len(result.Plans))
	}
	item := result.Plans[0].Item
	if item.OriginalUnitPrice.String() != "100.00" {
		t.Fatalf("expected original unit price 100.00, got %s", item.OriginalUnitPrice.String())
	}
	if item.OriginalTotalPrice.String() != "200.00" {
		t.Fatalf("expected original total price 200.00, got %s", item.OriginalTotalPrice.String())
	}
	if item.UnitPrice.String() != "72.00" {
		t.Fatalf("expected final unit price 72.00, got %s", item.UnitPrice.String())
	}
	if item.TotalPrice.String() != "144.00" {
		t.Fatalf("expected final total price 144.00, got %s", item.TotalPrice.String())
	}
}

func TestBuildOrderResultRejectsProductMaxPurchaseQuantityExceeded(t *testing.T) {
	assertBuildOrderResultRejectsPurchaseQuantity(t, orderPurchaseQuantityLimitFixture{
		dsnPrefix:       "order_service_purchase_limit",
		categorySlug:    "test-category-limit",
		productSlug:     "test-product-limit",
		maxQuantity:     2,
		requestQuantity: 3,
		expectedErr:     ErrProductMaxPurchaseExceeded,
	})
}

func TestBuildOrderResultRejectsProductMinPurchaseQuantityNotMet(t *testing.T) {
	assertBuildOrderResultRejectsPurchaseQuantity(t, orderPurchaseQuantityLimitFixture{
		dsnPrefix:       "order_service_purchase_min",
		categorySlug:    "test-category-min-limit",
		productSlug:     "test-product-min-limit",
		minQuantity:     3,
		requestQuantity: 2,
		expectedErr:     ErrProductMinPurchaseNotMet,
	})
}

func TestBuildOrderResultOriginalAmountBeforePromotion(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_promo_original_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductSKU{}, &models.Promotion{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now()
	category := models.Category{
		Slug:      "test-category-original",
		NameJSON:  jsonmap.JSON{"zh-CN": "测试分类"},
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	product := models.Product{
		CategoryID:      category.ID,
		Slug:            "test-product-original",
		TitleJSON:       jsonmap.JSON{"zh-CN": "测试商品"},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.RequireFromString("59.90")),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeManual,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	sku := models.ProductSKU{
		ProductID:         product.ID,
		SKUCode:           models.DefaultSKUCode,
		PriceAmount:       models.NewMoneyFromDecimal(decimal.RequireFromString("59.90")),
		IsActive:          true,
		ManualStockTotal:  constants.ManualStockUnlimited,
		ManualStockLocked: 0,
		ManualStockSold:   0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	promotion := models.Promotion{
		Name:       "test-20-percent",
		ScopeType:  constants.ScopeTypeProduct,
		ScopeRefID: product.ID,
		Type:       constants.PromotionTypePercent,
		Value:      models.NewMoneyFromDecimal(decimal.NewFromInt(20)),
		MinAmount:  models.NewMoneyFromDecimal(decimal.Zero),
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := db.Create(&promotion).Error; err != nil {
		t.Fatalf("create promotion failed: %v", err)
	}

	svc := NewOrderService(OrderServiceOptions{
		ProductRepo:    repository.NewProductRepository(db),
		ProductSKURepo: repository.NewProductSKURepository(db),
		PromotionRepo:  promotiongormstore.New(db),
		ExpireMinutes:  15,
	})

	result, err := svc.buildOrderResult(orderCreateParams{
		UserID: 1,
		Items: []CreateOrderItem{
			{
				ProductID: product.ID,
				SKUID:     sku.ID,
				Quantity:  2,
			},
		},
	})
	if err != nil {
		t.Fatalf("buildOrderResult failed: %v", err)
	}

	expectedOriginal := decimal.RequireFromString("119.80")
	expectedPromotion := decimal.RequireFromString("23.96")
	expectedTotal := decimal.RequireFromString("95.84")

	if !result.OriginalAmount.Equal(expectedOriginal) {
		t.Fatalf("expected original amount %s, got: %s", expectedOriginal.String(), result.OriginalAmount.String())
	}
	if !result.PromotionDiscountAmount.Equal(expectedPromotion) {
		t.Fatalf("expected promotion discount amount %s, got: %s", expectedPromotion.String(), result.PromotionDiscountAmount.String())
	}
	if !result.DiscountAmount.Equal(decimal.Zero) {
		t.Fatalf("expected coupon discount amount 0, got: %s", result.DiscountAmount.String())
	}
	if !result.TotalAmount.Equal(expectedTotal) {
		t.Fatalf("expected total amount %s, got: %s", expectedTotal.String(), result.TotalAmount.String())
	}
}

func TestBuildOrderResultRejectsZeroTotalAmountAfterCoupon(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_coupon_zero_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductSKU{}, &models.Coupon{}, &models.CouponUsage{}, &models.Promotion{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now()
	category := models.Category{
		Slug:      "test-category-coupon",
		NameJSON:  jsonmap.JSON{"zh-CN": "测试分类"},
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	product := models.Product{
		CategoryID:      category.ID,
		Slug:            "test-product-coupon",
		TitleJSON:       jsonmap.JSON{"zh-CN": "测试商品"},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeManual,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	sku := models.ProductSKU{
		ProductID:         product.ID,
		SKUCode:           models.DefaultSKUCode,
		PriceAmount:       models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		IsActive:          true,
		ManualStockTotal:  constants.ManualStockUnlimited,
		ManualStockLocked: 0,
		ManualStockSold:   0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	coupon := models.Coupon{
		Code:        "FREE10",
		Type:        constants.CouponTypeFixed,
		Value:       models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		MinAmount:   models.NewMoneyFromDecimal(decimal.Zero),
		MaxDiscount: models.NewMoneyFromDecimal(decimal.Zero),
		ScopeType:   constants.ScopeTypeProduct,
		ScopeRefIDs: fmt.Sprintf("[%d]", product.ID),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&coupon).Error; err != nil {
		t.Fatalf("create coupon failed: %v", err)
	}

	svc := NewOrderService(OrderServiceOptions{
		ProductRepo:     repository.NewProductRepository(db),
		ProductSKURepo:  repository.NewProductSKURepository(db),
		CouponRepo:      coupongormstore.New(db),
		CouponUsageRepo: coupongormstore.NewUsageStore(db),
		PromotionRepo:   promotiongormstore.New(db),
		ExpireMinutes:   15,
	})

	_, err = svc.buildOrderResult(orderCreateParams{
		UserID:     1,
		CouponCode: "FREE10",
		Items: []CreateOrderItem{
			{
				ProductID: product.ID,
				SKUID:     sku.ID,
				Quantity:  1,
			},
		},
	})
	if !errors.Is(err, ErrInvalidOrderAmount) {
		t.Fatalf("expected invalid order amount, got: %v", err)
	}
}
