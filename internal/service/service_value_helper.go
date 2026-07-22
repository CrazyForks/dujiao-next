package service

import (
	"fmt"
	"strings"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"

	"github.com/shopspring/decimal"
)

// resolveServiceSiteCurrency 统一解析服务层使用的站点币种配置。
func resolveServiceSiteCurrency(settingService *settingsapp.Service) string {
	if settingService == nil {
		return constants.SiteCurrencyDefault
	}
	currency, err := settingService.GetSiteCurrency(constants.SiteCurrencyDefault)
	if err != nil {
		return constants.SiteCurrencyDefault
	}
	return currency
}

// resolveOrderPaymentExpireMinutes 统一解析订单支付超时分钟配置。
func resolveOrderPaymentExpireMinutes(settingService *settingsapp.Service, defaultMinutes int) int {
	if defaultMinutes <= 0 {
		defaultMinutes = 15
	}
	if settingService == nil {
		return defaultMinutes
	}
	minutes, err := settingService.GetOrderPaymentExpireMinutes(defaultMinutes)
	if err != nil {
		return defaultMinutes
	}
	if minutes <= 0 {
		return defaultMinutes
	}
	return minutes
}

// resolveProductOrderSKU 统一解析下单相关场景的 SKU 选择逻辑。
func resolveProductOrderSKU(productSKURepo repository.ProductSKURepository, product *models.Product, rawSKUID uint) (*models.ProductSKU, error) {
	if product == nil || product.ID == 0 {
		return nil, ErrProductNotAvailable
	}
	if productSKURepo == nil {
		return nil, ErrProductSKUInvalid
	}

	if rawSKUID > 0 {
		sku, err := productSKURepo.GetByID(rawSKUID)
		if err != nil {
			return nil, err
		}
		if sku == nil || sku.ProductID != product.ID || !sku.IsActive {
			return nil, ErrProductSKUInvalid
		}
		return sku, nil
	}

	// 兼容窗口：无 sku_id 时仅允许“商品存在且仅存在一个启用 SKU”自动回退。
	activeSKUs, err := productSKURepo.ListByProduct(product.ID, true)
	if err != nil {
		return nil, err
	}
	if len(activeSKUs) == 1 {
		return &activeSKUs[0], nil
	}
	if len(activeSKUs) == 0 {
		return nil, ErrProductSKUInvalid
	}
	return nil, ErrProductSKURequired
}

// buildOrderItemKey 构建商品与 SKU 的组合键。
func buildOrderItemKey(productID, skuID uint) string {
	return fmt.Sprintf("%d:%d", productID, skuID)
}

// normalizeAffiliateCode 统一归一化推广码输入。
func normalizeAffiliateCode(raw string) string {
	code := strings.TrimSpace(raw)
	if code == "" {
		return ""
	}
	if len(code) > 32 {
		return code[:32]
	}
	return code
}

// normalizeOrderAmount 归一化金额精度与下限。
func normalizeOrderAmount(amount decimal.Decimal) decimal.Decimal {
	normalized := amount.Round(2)
	if normalized.LessThan(decimal.Zero) {
		return decimal.Zero
	}
	return normalized
}
