package catalogwiring

import (
	promotionmodule "github.com/dujiao-next/internal/modules/promotion"
	"github.com/dujiao-next/internal/provider"
	catalogtransport "github.com/dujiao-next/internal/transport/http/catalog"
)

func NewPublicHandler(c *provider.Container) *catalogtransport.PublicHandler {
	var promotions catalogtransport.ProductPromotionDecorator
	if c.PromotionRepo != nil {
		promotions = promotionmodule.NewService(c.PromotionRepo)
	}
	return catalogtransport.NewPublicHandler(
		catalogPublicProductAdapter{products: c.ProductService, hidden: c.ResellerRepo},
		c.CategoryService,
		c.ResellerPricingResolver,
		promotions,
		c.MemberLevelService,
		c.ProductMappingRepo,
		c.SKUMappingRepo,
		c.ContentPostService,
	)
}
