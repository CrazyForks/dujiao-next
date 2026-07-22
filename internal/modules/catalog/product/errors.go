package product

import "errors"

// 商品后台/写入相关哨兵。Service 兼容层与 HTTP transport 共用同一 identity。
var (
	ErrNotFound                     = errors.New("not found")
	ErrSlugExists                   = errors.New("slug exists")
	ErrProductCategoryInvalid       = errors.New("product category invalid")
	ErrProductPriceInvalid          = errors.New("product price invalid")
	ErrProductPurchaseInvalid       = errors.New("product purchase invalid")
	ErrProductPurchaseLimitInvalid  = errors.New("product purchase limit invalid")
	ErrProductStockDisplayInvalid   = errors.New("product stock display invalid")
	ErrFulfillmentInvalid           = errors.New("fulfillment invalid")
	ErrManualStockInvalid           = errors.New("manual stock invalid")
	ErrProductSKUInvalid            = errors.New("product sku invalid")
	ErrProductSKUHasCardSecretStock = errors.New("product sku has card secret stock")
	ErrProductHasStock              = errors.New("product has stock")
	ErrProductHasOrderRecord        = errors.New("product has order record")
	ErrResellerProductNotListed     = errors.New("reseller product not listed")
)
