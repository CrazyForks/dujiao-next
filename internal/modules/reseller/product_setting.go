package reseller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/dujiao-next/internal/models"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/shopspring/decimal"
)

// ProductSettingService 分销商品改价/上架配置用例。
type ProductSettingService struct {
	store       ProductSettingStore
	productRepo catalogproduct.Repository
}

func NewProductSettingService(store ProductSettingStore, productRepo catalogproduct.Repository) *ProductSettingService {
	return &ProductSettingService{store: store, productRepo: productRepo}
}

type ProductSettingInput struct {
	SKUID             uint
	IsListed          bool
	PricingMode       string
	MarkupPercent     decimal.Decimal
	FixedMarkupAmount decimal.Decimal
	FixedPriceAmount  decimal.Decimal
	SortOrder         int
}

type ProductSettingSaveInput struct {
	Settings []ProductSettingInput
}

type ProductSettingUserListInput struct {
	Page       int
	PageSize   int
	Keyword    string
	CategoryID uint
	Configured string
	Listed     string
}

type ProductSettingAdminListInput struct {
	Page        int
	PageSize    int
	ResellerID  uint
	UserID      uint
	ProductID   uint
	Keyword     string
	PricingMode string
	Configured  string
	Listed      string
}

type ProductSettingDetail struct {
	Profile          *models.ResellerProfile
	Product          models.Product
	Settings         []models.ResellerProductSetting
	EffectiveBySKUID map[uint]decimal.Decimal
	RuleBySKUID      map[uint]string
}

type ProductSettingListRow struct {
	Profile          *models.ResellerProfile
	Product          models.Product
	Settings         []models.ResellerProductSetting
	EffectiveBySKUID map[uint]decimal.Decimal
	RuleBySKUID      map[uint]string
}

// ProductSettingPreviewItem 表示某个商品级（SKUID=0）或 SKU 级规则在「拟用配置」下的预览结果。
// 复用与保存/下单完全一致的 ResolveUnitAmount + ValidateUnitAmount，确保预览价与实际成交价零分歧。
type ProductSettingPreviewItem struct {
	SKUID          uint
	IsListed       bool
	BasePrice      decimal.Decimal
	EffectivePrice decimal.Decimal
	Valid          bool
	ErrorCode      string
}

func (s *ProductSettingService) ListUserProductSettings(userID uint, input ProductSettingUserListInput) ([]ProductSettingListRow, int64, error) {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return nil, 0, err
	}
	rows, total, err := s.store.ListProductsWithSettings(ProductSettingListFilter{
		Page:       input.Page,
		PageSize:   input.PageSize,
		ResellerID: profile.ID,
		CategoryID: input.CategoryID,
		Keyword:    input.Keyword,
		Configured: input.Configured,
		Listed:     input.Listed,
		OnlyActive: true,
	})
	if err != nil {
		return nil, 0, err
	}
	return s.decorateRows(profile, rows), total, nil
}

func (s *ProductSettingService) GetUserProductSetting(userID, productID uint) (*ProductSettingDetail, error) {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return nil, err
	}
	return s.getDetail(profile, productID)
}

// PreviewUserProductSettings 在不落库的前提下，按用户拟用的定价规则计算各 SKU 的预计生效价与校验结果。
func (s *ProductSettingService) PreviewUserProductSettings(userID, productID uint, input ProductSettingSaveInput) ([]ProductSettingPreviewItem, error) {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return nil, err
	}
	return s.previewSettings(profile, productID, input)
}

func (s *ProductSettingService) SaveUserProductSettings(userID, productID uint, input ProductSettingSaveInput) (*ProductSettingDetail, error) {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return nil, err
	}
	if err := s.saveSettings(profile, productID, input); err != nil {
		return nil, err
	}
	return s.getDetail(profile, productID)
}

func (s *ProductSettingService) ResetUserProductSetting(userID, productID, skuID uint) error {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return err
	}
	return s.store.DeleteSetting(profile.ID, productID, skuID)
}

func (s *ProductSettingService) ListAdminSettings(input ProductSettingAdminListInput) ([]models.ResellerProductSetting, int64, error) {
	return s.store.ListAdminSettings(ProductSettingAdminListFilter(input))
}

func (s *ProductSettingService) SummarizeAdminSettings(resellerID uint) (ProductSettingSummary, error) {
	if s == nil || s.store == nil || resellerID == 0 {
		return ProductSettingSummary{}, nil
	}
	return s.store.SummarizeByResellerID(resellerID)
}

func (s *ProductSettingService) GetAdminProductSetting(resellerID, productID uint) (*ProductSettingDetail, error) {
	profile, err := s.requireActiveProfileByID(resellerID)
	if err != nil {
		return nil, err
	}
	return s.getDetail(profile, productID)
}

// PreviewAdminProductSettings 在不落库的前提下，按管理员拟用的定价规则计算各 SKU 的预计生效价与校验结果。
func (s *ProductSettingService) PreviewAdminProductSettings(resellerID, productID uint, input ProductSettingSaveInput) ([]ProductSettingPreviewItem, error) {
	profile, err := s.requireActiveProfileByID(resellerID)
	if err != nil {
		return nil, err
	}
	return s.previewSettings(profile, productID, input)
}

func (s *ProductSettingService) SaveAdminProductSettings(resellerID, productID uint, input ProductSettingSaveInput) (*ProductSettingDetail, error) {
	profile, err := s.requireActiveProfileByID(resellerID)
	if err != nil {
		return nil, err
	}
	if err := s.saveSettings(profile, productID, input); err != nil {
		return nil, err
	}
	return s.getDetail(profile, productID)
}

func (s *ProductSettingService) ResetAdminProductSetting(resellerID, productID, skuID uint) error {
	profile, err := s.requireActiveProfileByID(resellerID)
	if err != nil {
		return err
	}
	return s.store.DeleteSetting(profile.ID, productID, skuID)
}

func (s *ProductSettingService) requireActiveProfileByUser(userID uint) (*models.ResellerProfile, error) {
	if s == nil || s.store == nil || s.productRepo == nil || userID == 0 {
		return nil, catalogproduct.ErrNotFound
	}
	profile, err := s.store.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrNotOpened
	}
	if profile.Status != models.ResellerProfileStatusActive {
		return nil, ErrProfileInactive
	}
	return profile, nil
}

func (s *ProductSettingService) requireActiveProfileByID(resellerID uint) (*models.ResellerProfile, error) {
	if s == nil || s.store == nil || s.productRepo == nil || resellerID == 0 {
		return nil, catalogproduct.ErrNotFound
	}
	profile, err := s.store.GetProfileByID(resellerID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, catalogproduct.ErrNotFound
	}
	if profile.Status != models.ResellerProfileStatusActive {
		return nil, ErrProfileInactive
	}
	return profile, nil
}

func (s *ProductSettingService) getDetail(profile *models.ResellerProfile, productID uint) (*ProductSettingDetail, error) {
	if profile == nil || productID == 0 {
		return nil, catalogproduct.ErrNotFound
	}
	row, err := s.store.GetProductWithSettings(profile.ID, productID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, catalogproduct.ErrNotFound
	}
	effective, rules, err := computeProductEffectivePrices(profile, row.Product, row.Settings)
	if err != nil {
		return nil, err
	}
	return &ProductSettingDetail{Profile: profile, Product: row.Product, Settings: row.Settings, EffectiveBySKUID: effective, RuleBySKUID: rules}, nil
}

func (s *ProductSettingService) decorateRows(profile *models.ResellerProfile, rows []ProductSettingProductRow) []ProductSettingListRow {
	out := make([]ProductSettingListRow, 0, len(rows))
	for _, row := range rows {
		effective, rules, _ := computeProductEffectivePrices(profile, row.Product, row.Settings)
		out = append(out, ProductSettingListRow{Profile: profile, Product: row.Product, Settings: row.Settings, EffectiveBySKUID: effective, RuleBySKUID: rules})
	}
	return out
}

func (s *ProductSettingService) saveSettings(profile *models.ResellerProfile, productID uint, input ProductSettingSaveInput) error {
	if profile == nil || productID == 0 {
		return catalogproduct.ErrNotFound
	}
	product, err := s.productRepo.GetAdminByID(strconv.FormatUint(uint64(productID), 10))
	if err != nil {
		return err
	}
	if product == nil {
		return catalogproduct.ErrNotFound
	}
	normalizedSettings := make([]models.ResellerProductSetting, 0, len(input.Settings))
	for _, item := range input.Settings {
		normalized, err := normalizeProductSettingInput(profile, product, item)
		if err != nil {
			return err
		}
		normalized.ResellerID = profile.ID
		normalized.ProductID = product.ID
		normalizedSettings = append(normalizedSettings, normalized)
	}
	return s.store.WithinTransaction(func(store ProductSettingStore) error {
		for _, setting := range normalizedSettings {
			if _, err := store.UpsertSetting(setting); err != nil {
				return err
			}
		}
		return nil
	})
}

func normalizeProductSettingInput(profile *models.ResellerProfile, product *models.Product, input ProductSettingInput) (models.ResellerProductSetting, error) {
	if product == nil {
		return models.ResellerProductSetting{}, catalogproduct.ErrNotFound
	}
	mode := strings.TrimSpace(input.PricingMode)
	if mode == "" {
		mode = models.ResellerPricingModeInherit
	}
	switch mode {
	case models.ResellerPricingModeInherit, models.ResellerPricingModeMarkupPercent, models.ResellerPricingModeFixedMarkup, models.ResellerPricingModeFixedPrice:
	default:
		return models.ResellerProductSetting{}, ErrPricingModeInvalid
	}
	setting := models.ResellerProductSetting{
		SKUID:             input.SKUID,
		IsListed:          input.IsListed,
		PricingMode:       mode,
		MarkupPercent:     models.NewMoneyFromDecimal(input.MarkupPercent.Round(2)),
		FixedMarkupAmount: models.NewMoneyFromDecimal(input.FixedMarkupAmount.Round(2)),
		FixedPriceAmount:  models.NewMoneyFromDecimal(input.FixedPriceAmount.Round(2)),
		SortOrder:         input.SortOrder,
	}
	if !setting.IsListed {
		return setting, nil
	}
	if input.SKUID > 0 {
		sku := findProductSKU(product.SKUs, input.SKUID)
		if sku == nil || !sku.IsActive {
			return models.ResellerProductSetting{}, catalogproduct.ErrProductSKUInvalid
		}
		price, _, err := ResolveUnitAmount(profile, nil, &setting, sku.PriceAmount.Decimal.Round(2))
		if err != nil {
			return models.ResellerProductSetting{}, err
		}
		if err := ValidateUnitAmount(profile, sku, sku.PriceAmount.Decimal.Round(2), price); err != nil {
			return models.ResellerProductSetting{}, err
		}
		return setting, nil
	}
	if len(product.SKUs) == 0 {
		basePrice := product.PriceAmount.Decimal.Round(2)
		price, _, err := ResolveUnitAmount(profile, &setting, nil, basePrice)
		if err != nil {
			return models.ResellerProductSetting{}, err
		}
		if err := ValidateUnitAmount(profile, nil, basePrice, price); err != nil {
			return models.ResellerProductSetting{}, err
		}
		costPrice := product.CostPriceAmount.Decimal.Round(2)
		if costPrice.GreaterThan(decimal.Zero) && price.LessThan(costPrice) {
			return models.ResellerProductSetting{}, ErrPriceBelowBase
		}
		return setting, nil
	}
	for i := range product.SKUs {
		sku := &product.SKUs[i]
		if !sku.IsActive {
			continue
		}
		price, _, err := ResolveUnitAmount(profile, &setting, nil, sku.PriceAmount.Decimal.Round(2))
		if err != nil {
			return models.ResellerProductSetting{}, err
		}
		if err := ValidateUnitAmount(profile, sku, sku.PriceAmount.Decimal.Round(2), price); err != nil {
			return models.ResellerProductSetting{}, err
		}
	}
	return setting, nil
}

func (s *ProductSettingService) previewSettings(profile *models.ResellerProfile, productID uint, input ProductSettingSaveInput) ([]ProductSettingPreviewItem, error) {
	if profile == nil || productID == 0 {
		return nil, catalogproduct.ErrNotFound
	}
	row, err := s.store.GetProductWithSettings(profile.ID, productID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, catalogproduct.ErrNotFound
	}
	product := row.Product

	var productSetting *models.ResellerProductSetting
	skuSettings := map[uint]*models.ResellerProductSetting{}
	for _, item := range input.Settings {
		setting, err := buildPreviewSetting(item)
		if err != nil {
			return nil, err
		}
		setting.ResellerID = profile.ID
		setting.ProductID = product.ID
		if item.SKUID == 0 {
			ps := setting
			productSetting = &ps
		} else {
			ss := setting
			skuSettings[item.SKUID] = &ss
		}
	}

	items := make([]ProductSettingPreviewItem, 0, len(product.SKUs)+1)

	productBase := product.PriceAmount.Decimal.Round(2)
	productItem := ProductSettingPreviewItem{
		SKUID:     0,
		IsListed:  productSetting == nil || productSetting.IsListed,
		BasePrice: productBase,
		Valid:     true,
	}
	if productSetting != nil && productSetting.IsListed {
		price, _, _ := ResolveUnitAmount(profile, productSetting, nil, productBase)
		productItem.EffectivePrice = price.Round(2)
		if len(product.SKUs) == 0 {
			perr := ValidateUnitAmount(profile, nil, productBase, price)
			if perr == nil {
				cost := product.CostPriceAmount.Decimal.Round(2)
				if cost.GreaterThan(decimal.Zero) && price.LessThan(cost) {
					perr = ErrPriceBelowBase
				}
			}
			productItem.Valid = perr == nil
			productItem.ErrorCode = previewErrorCode(perr)
		} else {
			productItem.Valid, productItem.ErrorCode = previewValidateProductRuleAcrossSKUs(profile, product, productSetting)
		}
	}
	items = append(items, productItem)

	for i := range product.SKUs {
		sku := &product.SKUs[i]
		if !sku.IsActive {
			continue
		}
		skuSetting := skuSettings[sku.ID]
		base := sku.PriceAmount.Decimal.Round(2)
		item := ProductSettingPreviewItem{SKUID: sku.ID, IsListed: true, BasePrice: base, Valid: true}
		if (productSetting != nil && !productSetting.IsListed) || (skuSetting != nil && !skuSetting.IsListed) {
			item.IsListed = false
			items = append(items, item)
			continue
		}
		price, _, perr := ResolveUnitAmount(profile, productSetting, skuSetting, base)
		if perr == nil {
			perr = ValidateUnitAmount(profile, sku, base, price)
		}
		item.EffectivePrice = price.Round(2)
		item.Valid = perr == nil
		item.ErrorCode = previewErrorCode(perr)
		items = append(items, item)
	}
	return items, nil
}

func buildPreviewSetting(input ProductSettingInput) (models.ResellerProductSetting, error) {
	mode := strings.TrimSpace(input.PricingMode)
	if mode == "" {
		mode = models.ResellerPricingModeInherit
	}
	switch mode {
	case models.ResellerPricingModeInherit, models.ResellerPricingModeMarkupPercent, models.ResellerPricingModeFixedMarkup, models.ResellerPricingModeFixedPrice:
	default:
		return models.ResellerProductSetting{}, ErrPricingModeInvalid
	}
	return models.ResellerProductSetting{
		SKUID:             input.SKUID,
		IsListed:          input.IsListed,
		PricingMode:       mode,
		MarkupPercent:     models.NewMoneyFromDecimal(input.MarkupPercent.Round(2)),
		FixedMarkupAmount: models.NewMoneyFromDecimal(input.FixedMarkupAmount.Round(2)),
		FixedPriceAmount:  models.NewMoneyFromDecimal(input.FixedPriceAmount.Round(2)),
		SortOrder:         input.SortOrder,
	}, nil
}

func previewValidateProductRuleAcrossSKUs(profile *models.ResellerProfile, product models.Product, productSetting *models.ResellerProductSetting) (bool, string) {
	for i := range product.SKUs {
		sku := &product.SKUs[i]
		if !sku.IsActive {
			continue
		}
		base := sku.PriceAmount.Decimal.Round(2)
		price, _, err := ResolveUnitAmount(profile, productSetting, nil, base)
		if err == nil {
			err = ValidateUnitAmount(profile, sku, base, price)
		}
		if err != nil {
			return false, previewErrorCode(err)
		}
	}
	return true, ""
}

func previewErrorCode(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrMarkupExceeded):
		return "markup_exceeded"
	default:
		return "price_invalid"
	}
}

func computeProductEffectivePrices(profile *models.ResellerProfile, product models.Product, settings []models.ResellerProductSetting) (map[uint]decimal.Decimal, map[uint]string, error) {
	effective := map[uint]decimal.Decimal{}
	rules := map[uint]string{}
	byProduct, bySKU := indexProductSettings(settings)
	productSetting := byProduct[product.ID]
	if productSetting != nil && productSetting.IsListed {
		price, rule, err := ResolveUnitAmount(profile, productSetting, nil, product.PriceAmount.Decimal.Round(2))
		if err != nil {
			return effective, rules, err
		}
		effective[0] = price.Round(2)
		rules[0] = rule.Source
	}
	for i := range product.SKUs {
		sku := &product.SKUs[i]
		if !sku.IsActive {
			continue
		}
		skuSetting := bySKU[productSettingKey{productID: product.ID, skuID: sku.ID}]
		if productSetting != nil && !productSetting.IsListed {
			continue
		}
		if skuSetting != nil && !skuSetting.IsListed {
			continue
		}
		price, rule, err := ResolveUnitAmount(profile, productSetting, skuSetting, sku.PriceAmount.Decimal.Round(2))
		if err != nil {
			return effective, rules, err
		}
		effective[sku.ID] = price.Round(2)
		rules[sku.ID] = rule.Source
	}
	return effective, rules, nil
}

type productSettingKey struct {
	productID uint
	skuID     uint
}

func indexProductSettings(settings []models.ResellerProductSetting) (map[uint]*models.ResellerProductSetting, map[productSettingKey]*models.ResellerProductSetting) {
	byProduct := make(map[uint]*models.ResellerProductSetting)
	bySKU := make(map[productSettingKey]*models.ResellerProductSetting)
	for i := range settings {
		setting := settings[i]
		row := setting
		if setting.SKUID == 0 {
			byProduct[setting.ProductID] = &row
			continue
		}
		bySKU[productSettingKey{productID: setting.ProductID, skuID: setting.SKUID}] = &row
	}
	return byProduct, bySKU
}

func findProductSKU(items []models.ProductSKU, skuID uint) *models.ProductSKU {
	for i := range items {
		if items[i].ID == skuID {
			return &items[i]
		}
	}
	return nil
}
