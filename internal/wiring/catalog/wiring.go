package catalogwiring

import (
	"github.com/dujiao-next/internal/models"
	productapplication "github.com/dujiao-next/internal/modules/catalog/product/application"
	"github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/service"
)

// catalogPublicProductAdapter 将商品服务与分销隐藏商品端口适配为公开查询接口。
type catalogPublicProductAdapter struct {
	products *service.ProductService
	hidden   productapplication.HiddenProductRepository
}

func (a catalogPublicProductAdapter) ListPublicForTenant(
	tenant reseller.TenantContext,
	categoryID, search string,
	page, pageSize int,
) ([]models.Product, int64, error) {
	return a.products.ListPublicForTenant(tenant, a.hidden, categoryID, search, page, pageSize)
}

func (a catalogPublicProductAdapter) GetPublicBySlugForTenant(
	tenant reseller.TenantContext,
	slug string,
) (*models.Product, error) {
	return a.products.GetPublicBySlugForTenant(tenant, a.hidden, slug)
}

func (a catalogPublicProductAdapter) ApplyAutoStockCounts(products []models.Product) error {
	return a.products.ApplyAutoStockCounts(products)
}
