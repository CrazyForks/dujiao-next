package service

import (
	"github.com/dujiao-next/internal/modules/catalog"
	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	productapplication "github.com/dujiao-next/internal/modules/catalog/product/application"
	productadmin "github.com/dujiao-next/internal/modules/catalog/product/application/admin"
	productwrite "github.com/dujiao-next/internal/modules/catalog/product/application/write"
	"github.com/dujiao-next/internal/repository"
)

// ProductService 商品业务服务
type ProductService struct {
	*productapplication.Service
	*productadmin.AdminService
	*productwrite.WriteService
}

// NewProductService 创建商品服务
func NewProductService(
	repo repository.ProductRepository,
	productSKURepo repository.ProductSKURepository,
	cardSecretRepo repository.CardSecretRepository,
	cardSecretBatchRepo repository.CardSecretBatchRepository,
	categoryRepo catalog.CategoryRepository,
	memberLevelPriceRepo MemberLevelPriceCleaner,
	cartRepo repository.CartRepository,
	productMappingRepo catalogmapping.MappingRepository,
	orderRepo repository.OrderRepository,
	paymentChannelRepo repository.PaymentChannelRepository,
) *ProductService {
	return &ProductService{
		Service: productapplication.NewService(productapplication.Options{
			Products:                      repo,
			Categories:                    categoryRepo,
			Stock:                         cardSecretRepo,
			NotFoundError:                 ErrNotFound,
			ResellerProductNotListedError: ErrResellerProductNotListed,
		}),
		AdminService: productadmin.NewAdminService(productadmin.Options{
			Products:     repo,
			Categories:   categoryRepo,
			CardSecrets:  cardSecretRepo,
			Orders:       orderRepo,
			Transactions: newProductAdminUnitOfWork(repo, productSKURepo, cardSecretRepo, cardSecretBatchRepo, memberLevelPriceRepo, cartRepo, productMappingRepo),
			Errors: productadmin.ErrorSet{
				NotFound:               ErrNotFound,
				ProductCategoryInvalid: ErrProductCategoryInvalid,
				ProductHasStock:        ErrProductHasStock,
				ProductHasOrderRecord:  ErrProductHasOrderRecord,
			},
		}),
		WriteService: productwrite.NewWriteService(productwrite.Options{
			Products:        repo,
			SKUs:            productSKURepo,
			Categories:      categoryRepo,
			PaymentChannels: paymentChannelRepo,
			Transactions:    newProductWriteUnitOfWork(repo, productSKURepo, cardSecretRepo),
			Errors: productwrite.ErrorSet{
				NotFound:                     ErrNotFound,
				SlugExists:                   ErrSlugExists,
				ProductCategoryInvalid:       ErrProductCategoryInvalid,
				ProductPurchaseInvalid:       ErrProductPurchaseInvalid,
				FulfillmentInvalid:           ErrFulfillmentInvalid,
				ProductPriceInvalid:          ErrProductPriceInvalid,
				ManualStockInvalid:           ErrManualStockInvalid,
				ProductPurchaseLimitInvalid:  ErrProductPurchaseLimitInvalid,
				ProductStockDisplayInvalid:   ErrProductStockDisplayInvalid,
				ProductSKUInvalid:            ErrProductSKUInvalid,
				ProductSKUHasCardSecretStock: ErrProductSKUHasCardSecretStock,
			},
		}),
	}
}

type CreateProductInput = productwrite.CreateProductInput
type WholesalePriceInput = productwrite.WholesalePriceInput
type ProductSKUInput = productwrite.ProductSKUInput
