package service

import (
	"errors"
	"testing"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"

	categorydomain "github.com/dujiao-next/internal/modules/catalog/category/domain"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/shopspring/decimal"
)

func TestProductServiceCreateRejectsParentCategoryWithChildren(t *testing.T) {
	svc, db := newProductServiceForTest(t)

	parent := categorydomain.Category{
		Slug:     "games",
		NameJSON: jsonmap.JSON{"zh-CN": "games"},
	}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("create parent category failed: %v", err)
	}
	child := categorydomain.Category{
		ParentID: parent.ID,
		Slug:     "steam",
		NameJSON: jsonmap.JSON{"zh-CN": "steam"},
	}
	if err := db.Create(&child).Error; err != nil {
		t.Fatalf("create child category failed: %v", err)
	}

	_, err := svc.Create(CreateProductInput{
		CategoryID:      parent.ID,
		Slug:            "invalid-parent-product",
		TitleJSON:       map[string]interface{}{"zh-CN": "invalid-parent-product"},
		PriceAmount:     decimal.NewFromInt(10),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeManual,
		ManualStockTotal: func() *int {
			value := 1
			return &value
		}(),
	})
	if err != ErrProductCategoryInvalid {
		t.Fatalf("expected ErrProductCategoryInvalid, got %v", err)
	}
}

func TestProductServiceCreateFiltersUnavailablePaymentChannels(t *testing.T) {
	svc, db := newProductServiceForTest(t)

	category := categorydomain.Category{
		Slug:     "payment-channel-category",
		NameJSON: jsonmap.JSON{"zh-CN": "payment-channel-category"},
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	activeChannel := createProductTestPaymentChannel(t, db, "Active", true, false)
	inactiveChannel := createProductTestPaymentChannel(t, db, "Inactive", false, false)
	deletedChannel := createProductTestPaymentChannel(t, db, "Deleted", true, true)

	product, err := svc.Create(CreateProductInput{
		CategoryID:        category.ID,
		Slug:              "payment-channel-create",
		TitleJSON:         map[string]interface{}{"zh-CN": "payment-channel-create"},
		PriceAmount:       decimal.NewFromInt(10),
		PurchaseType:      constants.ProductPurchaseMember,
		FulfillmentType:   constants.FulfillmentTypeAuto,
		PaymentChannelIDs: []uint{deletedChannel.ID, inactiveChannel.ID, activeChannel.ID},
		IsActive: func() *bool {
			value := true
			return &value
		}(),
	})
	if err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	got := DecodeChannelIDs(product.PaymentChannelIDs)
	if len(got) != 1 || got[0] != activeChannel.ID {
		t.Fatalf("expected only active payment channel %d, got %v", activeChannel.ID, got)
	}
}

func TestProductServiceCreateRejectsInvalidPurchaseLimits(t *testing.T) {
	svc, db := newProductServiceForTest(t)

	cat := categorydomain.Category{Slug: "test-purchase-limit", NameJSON: jsonmap.JSON{"zh-CN": "test"}}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	intPtr := func(v int) *int { return &v }

	_, err := svc.Create(CreateProductInput{
		CategoryID:          cat.ID,
		Slug:                "invalid-limit-product",
		TitleJSON:           map[string]interface{}{"zh-CN": "invalid-limit-product"},
		PriceAmount:         decimal.NewFromInt(10),
		PurchaseType:        constants.ProductPurchaseMember,
		FulfillmentType:     constants.FulfillmentTypeManual,
		ManualStockTotal:    intPtr(1),
		MinPurchaseQuantity: intPtr(10),
		MaxPurchaseQuantity: intPtr(5),
	})
	if !errors.Is(err, ErrProductPurchaseLimitInvalid) {
		t.Fatalf("expected ErrProductPurchaseLimitInvalid, got %v", err)
	}
}

func TestProductServiceCreateRollsBackProductAndSKUWhenWholesaleValidationFails(t *testing.T) {
	svc, db := newProductServiceForTest(t)
	category := categorydomain.Category{
		Slug:     "write-rollback-category",
		NameJSON: jsonmap.JSON{"zh-CN": "write-rollback-category"},
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	_, err := svc.Create(CreateProductInput{
		CategoryID:      category.ID,
		Slug:            "write-rollback-product",
		TitleJSON:       map[string]interface{}{"zh-CN": "write-rollback-product"},
		PriceAmount:     decimal.NewFromInt(10),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeAuto,
		WholesalePrices: &[]WholesalePriceInput{{
			MinQuantity: 0,
			UnitPrice:   decimal.NewFromInt(8),
		}},
	})
	if !errors.Is(err, ErrWholesalePriceInvalid) {
		t.Fatalf("expected ErrWholesalePriceInvalid, got %v", err)
	}

	var productCount int64
	if err := db.Model(&productdomain.Product{}).Where("slug = ?", "write-rollback-product").Count(&productCount).Error; err != nil {
		t.Fatalf("count products: %v", err)
	}
	if productCount != 0 {
		t.Fatalf("transaction must roll back product, got %d rows", productCount)
	}
	var skuCount int64
	if err := db.Model(&productdomain.ProductSKU{}).Count(&skuCount).Error; err != nil {
		t.Fatalf("count product SKUs: %v", err)
	}
	if skuCount != 0 {
		t.Fatalf("transaction must roll back default SKU, got %d rows", skuCount)
	}
}
