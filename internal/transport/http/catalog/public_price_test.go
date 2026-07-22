package cataloghttp

import (
	"errors"
	"testing"

	"github.com/dujiao-next/internal/models"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/dujiao-next/internal/modules/reseller"
	"github.com/shopspring/decimal"
)

type stubResellerDisplayPricer struct {
	result *reseller.DisplayPriceResult
	err    error
}

func (s stubResellerDisplayPricer) LoadDisplayPricingBatch(tenant reseller.TenantContext, products []models.Product) (*reseller.DisplayPricingBatch, error) {
	return &reseller.DisplayPricingBatch{Tenant: tenant}, nil
}

func (s stubResellerDisplayPricer) ResolveDisplayPrices(tenant reseller.TenantContext, product *models.Product, batch *reseller.DisplayPricingBatch) (*reseller.DisplayPriceResult, error) {
	return s.result, s.err
}

func TestDecoratePublicProductDisplayPricePrefersFirstActiveSKU(t *testing.T) {
	h := &PublicHandler{}
	product := &models.Product{
		ID:          1,
		PriceAmount: models.NewMoneyFromDecimal(decimal.RequireFromString("59.90")),
		SKUs: []models.ProductSKU{
			{
				ID:          11,
				IsActive:    true,
				SortOrder:   100,
				PriceAmount: models.NewMoneyFromDecimal(decimal.RequireFromString("89.90")),
			},
			{
				ID:          12,
				IsActive:    true,
				SortOrder:   10,
				PriceAmount: models.NewMoneyFromDecimal(decimal.RequireFromString("49.90")),
			},
		},
	}

	item, err := h.decoratePublicProduct(product, nil)
	if err != nil {
		t.Fatalf("decoratePublicProduct failed: %v", err)
	}
	expected := decimal.RequireFromString("89.90")
	if !item.PriceAmount.Decimal.Equal(expected) {
		t.Fatalf("expected display price %s, got: %s", expected.String(), item.PriceAmount.String())
	}
}

func TestDecoratePublicProductForTenantUsesResellerPricesAndHidesMainDiscounts(t *testing.T) {
	product := &models.Product{
		ID:          1,
		Slug:        "reseller-display",
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		WholesalePrices: models.WholesalePriceTiers{
			{MinQuantity: 5, UnitPrice: models.NewMoneyFromDecimal(decimal.NewFromInt(80))},
		},
		SKUs: []models.ProductSKU{
			{
				ID:          11,
				ProductID:   1,
				SKUCode:     "VISIBLE",
				IsActive:    true,
				SortOrder:   100,
				PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
			},
			{
				ID:          12,
				ProductID:   1,
				SKUCode:     "HIDDEN",
				IsActive:    true,
				SortOrder:   10,
				PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(90)),
			},
		},
	}
	h := &PublicHandler{pricer: stubResellerDisplayPricer{result: &reseller.DisplayPriceResult{
		Visible:      true,
		ProductID:    1,
		DisplaySKUID: 11,
		DisplayPrice: models.NewMoneyFromDecimal(decimal.NewFromInt(130)),
		SKUPrices: map[uint]models.Money{
			11: models.NewMoneyFromDecimal(decimal.NewFromInt(130)),
		},
		HiddenSKUIDs: map[uint]bool{12: true},
	}}}
	tenant := reseller.ResellerTenantContext("shop.example.test", 10, 99, "shop.example.test")

	item, err := h.decoratePublicProductForTenant(product, nil, tenant, &reseller.DisplayPricingBatch{})
	if err != nil {
		t.Fatalf("decoratePublicProductForTenant failed: %v", err)
	}
	if !item.PriceAmount.Decimal.Equal(decimal.NewFromInt(130)) {
		t.Fatalf("expected reseller product price 130, got %s", item.PriceAmount.String())
	}
	if len(item.SKUs) != 1 || item.SKUs[0].ID != 11 {
		t.Fatalf("expected only visible sku 11, got %+v", item.SKUs)
	}
	if !item.SKUs[0].PriceAmount.Decimal.Equal(decimal.NewFromInt(130)) {
		t.Fatalf("expected reseller sku price 130, got %s", item.SKUs[0].PriceAmount.String())
	}
	if item.PromotionPriceAmount != nil || item.PromotionID != nil || len(item.PromotionRules) > 0 {
		t.Fatalf("reseller display must not expose main promotion fields: %+v", item)
	}
	if len(item.WholesalePrices) > 0 {
		t.Fatalf("reseller display must not expose main wholesale prices: %+v", item.WholesalePrices)
	}
}

func TestDecoratePublicProductForTenantHiddenProduct(t *testing.T) {
	h := &PublicHandler{pricer: stubResellerDisplayPricer{result: &reseller.DisplayPriceResult{Visible: false}}}
	tenant := reseller.ResellerTenantContext("shop.example.test", 10, 99, "shop.example.test")
	product := &models.Product{
		ID:          1,
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		SKUs: []models.ProductSKU{
			{ID: 11, ProductID: 1, IsActive: true, PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100))},
		},
	}

	_, err := h.decoratePublicProductForTenant(product, nil, tenant, &reseller.DisplayPricingBatch{})
	if !errors.Is(err, catalogproduct.ErrResellerProductNotListed) {
		t.Fatalf("expected ErrResellerProductNotListed, got %v", err)
	}
}

func TestDecoratePublicProductForTenantAllHiddenSKUsReturnsNotListed(t *testing.T) {
	h := &PublicHandler{pricer: stubResellerDisplayPricer{result: &reseller.DisplayPriceResult{
		Visible:      true,
		DisplayPrice: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		HiddenSKUIDs: map[uint]bool{11: true, 12: true},
	}}}
	tenant := reseller.ResellerTenantContext("shop.example.test", 10, 99, "shop.example.test")
	product := &models.Product{
		ID:          1,
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		SKUs: []models.ProductSKU{
			{ID: 11, ProductID: 1, IsActive: true, PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100))},
			{ID: 12, ProductID: 1, IsActive: true, PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(120))},
		},
	}

	_, err := h.decoratePublicProductForTenant(product, nil, tenant, &reseller.DisplayPricingBatch{})
	if !errors.Is(err, catalogproduct.ErrResellerProductNotListed) {
		t.Fatalf("expected ErrResellerProductNotListed, got %v", err)
	}
}

func TestDecoratePublicProductForTenantInvalidDisplayPricingIsHidden(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "below base", err: reseller.ErrPriceBelowBase},
		{name: "markup exceeds max", err: reseller.ErrMarkupExceeded},
		{name: "unknown pricing mode", err: reseller.ErrPricingModeInvalid},
		{name: "already not listed", err: catalogproduct.ErrResellerProductNotListed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &PublicHandler{pricer: stubResellerDisplayPricer{err: tt.err}}
			tenant := reseller.ResellerTenantContext("shop.example.test", 10, 99, "shop.example.test")
			product := &models.Product{
				ID:          1,
				PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
				SKUs: []models.ProductSKU{
					{ID: 11, ProductID: 1, IsActive: true, PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100))},
				},
			}
			_, err := h.decoratePublicProductForTenant(product, nil, tenant, &reseller.DisplayPricingBatch{})
			if !errors.Is(err, catalogproduct.ErrResellerProductNotListed) {
				t.Fatalf("expected ErrResellerProductNotListed, got %v", err)
			}
		})
	}
}
