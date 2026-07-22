package productapplication

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/cardsecret"
	"github.com/dujiao-next/internal/modules/catalog/product"
)

type productRepositoryStub struct {
	filter  product.ListFilter
	rows    []models.Product
	total   int64
	bySlug  *models.Product
	byAdmin *models.Product
}

func (stub *productRepositoryStub) List(filter product.ListFilter) ([]models.Product, int64, error) {
	stub.filter = filter
	return stub.rows, stub.total, nil
}

func (stub *productRepositoryStub) GetBySlug(string, bool) (*models.Product, error) {
	return stub.bySlug, nil
}

func (stub *productRepositoryStub) GetAdminByID(string) (*models.Product, error) {
	return stub.byAdmin, nil
}

type categoryRepositoryStub struct {
	byID map[string]*models.Category
	rows []models.Category
}

func (stub categoryRepositoryStub) GetByID(id string) (*models.Category, error) {
	return stub.byID[id], nil
}

func (stub categoryRepositoryStub) List() ([]models.Category, error) {
	return stub.rows, nil
}

type hiddenProductRepositoryStub struct {
	ids []uint
}

func (stub hiddenProductRepositoryStub) ListHiddenProductIDs(uint) ([]uint, error) {
	return stub.ids, nil
}

type stockCounterStub struct {
	counts []cardsecret.SKUStockCount
}

func (stub stockCounterStub) CountStockByProductIDs([]uint) ([]cardsecret.SKUStockCount, error) {
	return stub.counts, nil
}

func TestListPublicForTenantBuildsVisibilityFilterBeforePagination(t *testing.T) {
	products := &productRepositoryStub{total: 2}
	categories := categoryRepositoryStub{
		byID: map[string]*models.Category{
			"10": {ID: 10, IsActive: true},
		},
		rows: []models.Category{
			{ID: 10, IsActive: true},
			{ID: 11, ParentID: 10, IsActive: true},
			{ID: 12, ParentID: 10, IsActive: false},
		},
	}
	service := NewService(Options{Products: products, Categories: categories})
	resellerID := uint(9)

	_, total, err := service.ListPublicForTenant(
		TenantContext{ResellerID: &resellerID},
		hiddenProductRepositoryStub{ids: []uint{7, 8}},
		"10",
		"keyword",
		2,
		20,
	)
	if err != nil {
		t.Fatalf("ListPublicForTenant returned error: %v", err)
	}
	if total != 2 {
		t.Fatalf("total want 2 got %d", total)
	}
	if !reflect.DeepEqual(products.filter.CategoryIDs, []uint{10, 11}) {
		t.Fatalf("category ids want [10 11] got %v", products.filter.CategoryIDs)
	}
	if !reflect.DeepEqual(products.filter.ExcludeProductIDs, []uint{7, 8}) {
		t.Fatalf("excluded product ids want [7 8] got %v", products.filter.ExcludeProductIDs)
	}
	if products.filter.Page != 2 || products.filter.PageSize != 20 || !products.filter.OnlyActive || !products.filter.WithCategory {
		t.Fatalf("unexpected public filter: %#v", products.filter)
	}
}

func TestQueryServicePreservesInjectedCompatibilityErrors(t *testing.T) {
	notFound := errors.New("legacy not found")
	notListed := errors.New("legacy reseller product not listed")
	service := NewService(Options{
		Products:                      &productRepositoryStub{},
		NotFoundError:                 notFound,
		ResellerProductNotListedError: notListed,
	})

	if _, err := service.GetPublicBySlug("missing"); !errors.Is(err, notFound) {
		t.Fatalf("GetPublicBySlug want injected not-found error, got %v", err)
	}
	resellerID := uint(9)
	if _, _, err := service.ListPublicForTenant(TenantContext{ResellerID: &resellerID}, nil, "", "", 1, 20); !errors.Is(err, notListed) {
		t.Fatalf("ListPublicForTenant want injected not-listed error, got %v", err)
	}
}

func TestApplyAutoStockCountsAssignsLegacyStockToOneSKU(t *testing.T) {
	service := NewService(Options{Stock: stockCounterStub{counts: []cardsecret.SKUStockCount{
		{ProductID: 30, SKUID: 0, Status: models.CardSecretStatusAvailable, Total: 2},
		{ProductID: 30, SKUID: 101, Status: models.CardSecretStatusAvailable, Total: 3},
		{ProductID: 30, SKUID: 102, Status: models.CardSecretStatusAvailable, Total: 4},
	}}})
	products := []models.Product{{
		ID:              30,
		FulfillmentType: constants.FulfillmentTypeAuto,
		SKUs: []models.ProductSKU{
			{ID: 102, SKUCode: "SECOND", IsActive: true},
			{ID: 101, SKUCode: models.DefaultSKUCode, IsActive: true},
		},
	}}

	if err := service.ApplyAutoStockCounts(products); err != nil {
		t.Fatalf("ApplyAutoStockCounts returned error: %v", err)
	}
	if products[0].AutoStockAvailable != 9 {
		t.Fatalf("product available want 9 got %d", products[0].AutoStockAvailable)
	}
	if products[0].SKUs[0].AutoStockAvailable != 4 {
		t.Fatalf("secondary SKU available want 4 got %d", products[0].SKUs[0].AutoStockAvailable)
	}
	if products[0].SKUs[1].AutoStockAvailable != 5 {
		t.Fatalf("DEFAULT SKU available want 5 got %d", products[0].SKUs[1].AutoStockAvailable)
	}
}
