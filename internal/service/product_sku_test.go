package service

import (
	"strconv"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

func TestProductServiceUpdateRejectsDisablingAutoSKUWithCardSecretStock(t *testing.T) {
	svc, db := newProductServiceForTest(t)

	category := models.Category{
		Slug:     "auto-card-secret-category",
		NameJSON: models.JSON{"zh-CN": "auto-card-secret-category"},
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	product := models.Product{
		CategoryID:      category.ID,
		Slug:            "auto-card-secret-product",
		TitleJSON:       models.JSON{"zh-CN": "auto-card-secret-product"},
		PriceAmount:     models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeAuto,
		IsActive:        true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}

	stockSKU := models.ProductSKU{
		ProductID:      product.ID,
		SKUCode:        "SKU-STOCK",
		SpecValuesJSON: models.JSON{"zh-CN": "有库存"},
		PriceAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		IsActive:       true,
		SortOrder:      2,
	}
	spareSKU := models.ProductSKU{
		ProductID:      product.ID,
		SKUCode:        "SKU-SPARE",
		SpecValuesJSON: models.JSON{"zh-CN": "无库存"},
		PriceAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		IsActive:       true,
		SortOrder:      1,
	}
	if err := db.Create(&stockSKU).Error; err != nil {
		t.Fatalf("create stock sku failed: %v", err)
	}
	if err := db.Create(&spareSKU).Error; err != nil {
		t.Fatalf("create spare sku failed: %v", err)
	}

	insertCardSecrets(t, db, product.ID, stockSKU.ID, models.CardSecretStatusAvailable, 1)

	_, err := svc.Update(strconv.FormatUint(uint64(product.ID), 10), CreateProductInput{
		CategoryID:      category.ID,
		Slug:            product.Slug,
		TitleJSON:       map[string]interface{}{"zh-CN": "auto-card-secret-product"},
		PriceAmount:     decimal.NewFromInt(10),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeAuto,
		SKUs: []ProductSKUInput{
			{
				ID:             stockSKU.ID,
				SKUCode:        stockSKU.SKUCode,
				SpecValuesJSON: map[string]interface{}{"zh-CN": "有库存"},
				PriceAmount:    decimal.NewFromInt(10),
				IsActive: func() *bool {
					value := false
					return &value
				}(),
				SortOrder: 2,
			},
			{
				ID:             spareSKU.ID,
				SKUCode:        spareSKU.SKUCode,
				SpecValuesJSON: map[string]interface{}{"zh-CN": "无库存"},
				PriceAmount:    decimal.NewFromInt(10),
				IsActive: func() *bool {
					value := true
					return &value
				}(),
				SortOrder: 1,
			},
		},
		IsActive: func() *bool {
			value := true
			return &value
		}(),
	})
	if err != ErrProductSKUHasCardSecretStock {
		t.Fatalf("update product error want %v got %v", ErrProductSKUHasCardSecretStock, err)
	}
}
