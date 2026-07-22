package catalogwiring

import (
	productapplication "github.com/dujiao-next/internal/modules/catalog/product/application"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	"github.com/dujiao-next/internal/modules/reseller"
)

// catalogPublicProductAdapter 将商品服务与分销隐藏商品端口适配为公开查询接口。
type catalogPublicProductAdapter struct {
	products *productapplication.Service
	hidden   productapplication.HiddenProductRepository
}

func (a catalogPublicProductAdapter) ListPublicForTenant(
	tenant reseller.TenantContext,
	categoryID, search string,
	page, pageSize int,
) ([]productdomain.Product, int64, error) {
	return a.products.ListPublicForTenant(tenant, a.hidden, categoryID, search, page, pageSize)
}

func (a catalogPublicProductAdapter) GetPublicBySlugForTenant(
	tenant reseller.TenantContext,
	slug string,
) (*productdomain.Product, error) {
	return a.products.GetPublicBySlugForTenant(tenant, a.hidden, slug)
}

func (a catalogPublicProductAdapter) ApplyAutoStockCounts(products []productdomain.Product) error {
	return a.products.ApplyAutoStockCounts(products)
}
