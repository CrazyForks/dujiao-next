package productwrite

import (
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	"github.com/dujiao-next/internal/modules/catalog/product/manualform"

	"github.com/shopspring/decimal"
)

// Update 更新商品
func (s *WriteService) Update(id string, input CreateProductInput) (*models.Product, error) {
	priceAmount := input.PriceAmount.Round(2)
	if len(input.SKUs) == 0 && priceAmount.LessThanOrEqual(decimal.Zero) {
		return nil, s.errors.ProductPriceInvalid
	}
	product, err := s.products.GetByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, s.errors.NotFound
	}
	if err := productdomain.ValidateCategoryAssignment(s.categories, input.CategoryID, product.CategoryID, s.errors.ProductCategoryInvalid); err != nil {
		return nil, err
	}

	count, err := s.products.CountBySlug(input.Slug, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, s.errors.SlugExists
	}

	product.CategoryID = input.CategoryID
	product.Category = models.Category{}
	product.Slug = input.Slug
	product.SeoMetaJSON = models.JSON(input.SeoMetaJSON)
	product.TitleJSON = models.JSON(input.TitleJSON)
	product.DescriptionJSON = models.JSON(input.DescriptionJSON)
	product.ContentJSON = models.JSON(input.ContentJSON)
	product.InstructionsJSON = models.JSON(input.InstructionsJSON)
	product.ManualFormSchemaJSON = models.JSON{}
	product.PriceAmount = models.NewMoneyFromDecimal(priceAmount)
	product.SortOrder = input.SortOrder
	product.Images = models.StringArray(input.Images)
	product.Tags = models.StringArray(input.Tags)
	paymentChannelIDs, err := s.filterAvailablePaymentChannelIDs(input.PaymentChannelIDs)
	if err != nil {
		return nil, err
	}
	product.PaymentChannelIDs = encodeChannelIDs(paymentChannelIDs)
	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}
	if input.IsAffiliateEnabled != nil {
		product.IsAffiliateEnabled = *input.IsAffiliateEnabled
	}
	rawPurchaseType := strings.TrimSpace(input.PurchaseType)
	if rawPurchaseType == "" {
		rawPurchaseType = product.PurchaseType
	}
	purchaseType := productdomain.NormalizePurchaseType(rawPurchaseType)
	if purchaseType == "" {
		return nil, s.errors.ProductPurchaseInvalid
	}
	product.PurchaseType = purchaseType
	if input.MaxPurchaseQuantity != nil {
		product.MaxPurchaseQuantity = productdomain.NormalizePurchaseQuantityLimit(*input.MaxPurchaseQuantity)
	}
	if input.MinPurchaseQuantity != nil {
		product.MinPurchaseQuantity = productdomain.NormalizePurchaseQuantityLimit(*input.MinPurchaseQuantity)
	}
	if product.MinPurchaseQuantity > 0 && product.MaxPurchaseQuantity > 0 && product.MinPurchaseQuantity > product.MaxPurchaseQuantity {
		return nil, s.errors.ProductPurchaseLimitInvalid
	}
	stockDisplayMode := productdomain.NormalizeStockDisplayMode(input.StockDisplayMode)
	if stockDisplayMode == "" {
		return nil, s.errors.ProductStockDisplayInvalid
	}
	product.StockDisplayMode = stockDisplayMode
	rawFulfillmentType := strings.TrimSpace(input.FulfillmentType)
	if rawFulfillmentType == "" {
		rawFulfillmentType = product.FulfillmentType
	}
	fulfillmentType := productdomain.NormalizeFulfillmentType(rawFulfillmentType)
	if fulfillmentType == "" {
		return nil, s.errors.FulfillmentInvalid
	}
	// 对接商品的真实交付类型必须保持 upstream，后台返回的 auto/manual 仅用于展示。
	if product.IsMapped {
		fulfillmentType = constants.FulfillmentTypeUpstream
	}
	product.FulfillmentType = fulfillmentType
	if fulfillmentType == constants.FulfillmentTypeManual {
		normalizedSchemaJSON, err := manualform.NormalizeSchema(models.JSON(input.ManualFormSchemaJSON))
		if err != nil {
			return nil, err
		}
		product.ManualFormSchemaJSON = normalizedSchemaJSON
	}

	manualStockTotal := product.ManualStockTotal
	if input.ManualStockTotal != nil {
		manualStockTotal = *input.ManualStockTotal
	}
	if manualStockTotal < constants.ManualStockUnlimited {
		return nil, s.errors.ManualStockInvalid
	}

	var normalizedSKUs []normalizedProductSKU
	if len(input.SKUs) > 0 {
		if s.skus == nil {
			return nil, s.errors.ProductSKUInvalid
		}
		existingSKUs, listErr := s.skus.ListByProduct(product.ID, false)
		if listErr != nil {
			return nil, listErr
		}
		existingSKUMap := make(map[uint]models.ProductSKU, len(existingSKUs))
		for _, sku := range existingSKUs {
			existingSKUMap[sku.ID] = sku
		}
		var normalizeErr error
		normalizedSKUs, priceAmount, manualStockTotal, normalizeErr = s.normalizeProductSKUInputs(input.SKUs, fulfillmentType, existingSKUMap)
		if normalizeErr != nil {
			return nil, normalizeErr
		}
	}

	product.PriceAmount = models.NewMoneyFromDecimal(priceAmount)
	if len(normalizedSKUs) > 0 {
		product.CostPriceAmount = models.NewMoneyFromDecimal(minActiveCostPrice(normalizedSKUs))
	} else {
		product.CostPriceAmount = models.NewMoneyFromDecimal(input.CostPriceAmount.Round(2))
	}
	product.ManualStockTotal = manualStockTotal

	if err := s.transactions.WithinTransaction(func(repositories TransactionRepositories) error {
		productRepo := repositories.Products
		skuRepo := repositories.SKUs
		cardSecretRepo := repositories.CardSecrets
		if len(normalizedSKUs) > 0 {
			if err := s.applyProductSKUsWithStockGuard(skuRepo, cardSecretRepo, product.ID, fulfillmentType, normalizedSKUs); err != nil {
				return err
			}
		} else if err := s.syncSingleProductSKU(skuRepo, product.ID, priceAmount, product.CostPriceAmount.Decimal, product.ManualStockTotal, true); err != nil {
			return err
		}
		// 仅当请求显式携带批发价字段时才覆盖，省略字段（nil）保留原有配置，
		// 避免不关心批发价的局部更新静默清空已配阶梯。
		if input.WholesalePrices != nil {
			var skus []models.ProductSKU
			if skuRepo != nil {
				var err error
				skus, err = skuRepo.ListByProduct(product.ID, false)
				if err != nil {
					return err
				}
			}
			wholesalePrices, err := productdomain.NormalizeWholesalePricesForSKUs(*input.WholesalePrices, skus)
			if err != nil {
				return err
			}
			product.WholesalePrices = wholesalePrices
		}
		if err := productRepo.Update(product); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.products.GetByID(id)
}
