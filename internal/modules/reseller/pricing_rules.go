package reseller

import (
	"strings"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

const (
	RuleSourceSKU     = "sku"
	RuleSourceProduct = "product"
	RuleSourceProfile = "profile"
	RuleSourceInherit = "inherit"
)

// PricingRule 描述分销改价规则解析结果。
type PricingRule struct {
	Mode              string
	Source            string
	SettingID         *uint
	MarkupPercent     decimal.Decimal
	FixedMarkupAmount decimal.Decimal
	FixedPriceAmount  decimal.Decimal
}

// ResolveUnitAmount 按 SKU → 商品 → 资料默认加价 → 继承底价的优先级解析成交单价。
func ResolveUnitAmount(profile *models.ResellerProfile, productSetting *models.ResellerProductSetting, skuSetting *models.ResellerProductSetting, baseUnit decimal.Decimal) (decimal.Decimal, PricingRule, error) {
	if skuSetting != nil && strings.TrimSpace(skuSetting.PricingMode) != models.ResellerPricingModeInherit {
		return ApplyPricingRule(*skuSetting, RuleSourceSKU, baseUnit)
	}
	if productSetting != nil && strings.TrimSpace(productSetting.PricingMode) != models.ResellerPricingModeInherit {
		return ApplyPricingRule(*productSetting, RuleSourceProduct, baseUnit)
	}
	if profile != nil && profile.DefaultMarkupPercent.Decimal.GreaterThan(decimal.Zero) {
		rule := PricingRule{
			Mode:          models.ResellerPricingModeMarkupPercent,
			Source:        RuleSourceProfile,
			MarkupPercent: profile.DefaultMarkupPercent.Decimal.Round(2),
		}
		return ApplyMarkupPercent(baseUnit, rule.MarkupPercent), rule, nil
	}
	return baseUnit.Round(2), PricingRule{Mode: models.ResellerPricingModeInherit, Source: RuleSourceInherit}, nil
}

// ApplyPricingRule 按单条分销配置计算单价。
func ApplyPricingRule(setting models.ResellerProductSetting, source string, baseUnit decimal.Decimal) (decimal.Decimal, PricingRule, error) {
	settingID := setting.ID
	rule := PricingRule{
		Mode:              strings.TrimSpace(setting.PricingMode),
		Source:            source,
		SettingID:         &settingID,
		MarkupPercent:     setting.MarkupPercent.Decimal.Round(2),
		FixedMarkupAmount: setting.FixedMarkupAmount.Decimal.Round(2),
		FixedPriceAmount:  setting.FixedPriceAmount.Decimal.Round(2),
	}
	switch rule.Mode {
	case models.ResellerPricingModeMarkupPercent:
		return ApplyMarkupPercent(baseUnit, rule.MarkupPercent), rule, nil
	case models.ResellerPricingModeFixedMarkup:
		return baseUnit.Add(rule.FixedMarkupAmount).Round(2), rule, nil
	case models.ResellerPricingModeFixedPrice:
		return rule.FixedPriceAmount.Round(2), rule, nil
	case models.ResellerPricingModeInherit:
		return baseUnit.Round(2), rule, nil
	default:
		return decimal.Zero, rule, ErrPricingModeInvalid
	}
}

// ApplyMarkupPercent 按百分比加价。
func ApplyMarkupPercent(baseUnit decimal.Decimal, percent decimal.Decimal) decimal.Decimal {
	return baseUnit.Mul(decimal.NewFromInt(100).Add(percent)).Div(decimal.NewFromInt(100)).Round(2)
}

// ValidateUnitAmount 校验分销单价不低于底价/成本价，且不超过最大加价比例。
func ValidateUnitAmount(profile *models.ResellerProfile, sku *productdomain.ProductSKU, baseUnit decimal.Decimal, resellerUnit decimal.Decimal) error {
	baseUnit = baseUnit.Round(2)
	resellerUnit = resellerUnit.Round(2)
	if resellerUnit.LessThanOrEqual(decimal.Zero) || resellerUnit.LessThan(baseUnit) {
		return ErrPriceBelowBase
	}
	if sku != nil && sku.CostPriceAmount.Decimal.GreaterThan(decimal.Zero) && resellerUnit.LessThan(sku.CostPriceAmount.Decimal.Round(2)) {
		return ErrPriceBelowBase
	}
	if profile != nil && profile.MaxMarkupPercent.Decimal.GreaterThan(decimal.Zero) && baseUnit.GreaterThan(decimal.Zero) {
		implicit := resellerUnit.Sub(baseUnit).Div(baseUnit).Mul(decimal.NewFromInt(100)).Round(4)
		if implicit.GreaterThan(profile.MaxMarkupPercent.Decimal.Round(4)) {
			return ErrMarkupExceeded
		}
	}
	return nil
}
