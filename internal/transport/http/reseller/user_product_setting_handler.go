package resellerhttp

import (
	"strings"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

// UserProductSettingService 是用户中心分销商品配置端点所需的最小用例接口。
type UserProductSettingService interface {
	ListUserProductSettings(userID uint, input resellermodule.ProductSettingUserListInput) ([]resellermodule.ProductSettingListRow, int64, error)
	GetUserProductSetting(userID, productID uint) (*resellermodule.ProductSettingDetail, error)
	PreviewUserProductSettings(userID, productID uint, input resellermodule.ProductSettingSaveInput) ([]resellermodule.ProductSettingPreviewItem, error)
	SaveUserProductSettings(userID, productID uint, input resellermodule.ProductSettingSaveInput) (*resellermodule.ProductSettingDetail, error)
	ResetUserProductSetting(userID, productID, skuID uint) error
}

// UserProductSettingHandler 处理用户中心分销商品配置请求。
type UserProductSettingHandler struct {
	service UserProductSettingService
}

func NewUserProductSettingHandler(service UserProductSettingService) *UserProductSettingHandler {
	if service == nil {
		panic("reseller user product setting handler: service is nil")
	}
	return &UserProductSettingHandler{service: service}
}

type productSettingRequest struct {
	SKUID             uint   `json:"sku_id"`
	IsListed          bool   `json:"is_listed"`
	PricingMode       string `json:"pricing_mode"`
	MarkupPercent     string `json:"markup_percent"`
	FixedMarkupAmount string `json:"fixed_markup_amount"`
	FixedPriceAmount  string `json:"fixed_price_amount"`
	SortOrder         int    `json:"sort_order"`
}

type productSettingsUpdateRequest struct {
	Settings []productSettingRequest `json:"settings"`
}

func (req productSettingsUpdateRequest) toInput() (resellermodule.ProductSettingSaveInput, error) {
	input := resellermodule.ProductSettingSaveInput{Settings: make([]resellermodule.ProductSettingInput, 0, len(req.Settings))}
	for _, item := range req.Settings {
		markup, err := parseProductSettingDecimalField(item.MarkupPercent)
		if err != nil {
			return input, err
		}
		fixedMarkup, err := parseProductSettingDecimalField(item.FixedMarkupAmount)
		if err != nil {
			return input, err
		}
		fixedPrice, err := parseProductSettingDecimalField(item.FixedPriceAmount)
		if err != nil {
			return input, err
		}
		input.Settings = append(input.Settings, resellermodule.ProductSettingInput{
			SKUID:             item.SKUID,
			IsListed:          item.IsListed,
			PricingMode:       strings.TrimSpace(item.PricingMode),
			MarkupPercent:     markup,
			FixedMarkupAmount: fixedMarkup,
			FixedPriceAmount:  fixedPrice,
			SortOrder:         item.SortOrder,
		})
	}
	return input, nil
}

func parseProductSettingDecimalField(raw string) (decimal.Decimal, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(value)
}

// ListProductSettings 查询当前用户可配置的分销商品。
func (h *UserProductSettingHandler) ListProductSettings(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	page, pageSize := shared.ParsePagination(c)
	categoryID, _ := shared.ParseQueryUint(c.Query("category_id"), false)
	rows, total, err := h.service.ListUserProductSettings(uid, resellermodule.ProductSettingUserListInput{
		Page:       page,
		PageSize:   pageSize,
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		CategoryID: categoryID,
		Configured: strings.TrimSpace(c.Query("configured")),
		Listed:     strings.TrimSpace(c.Query("listed")),
	})
	if err != nil {
		respondUserProductSettingError(c, err, "error.user_fetch_failed")
		return
	}
	response.SuccessWithPage(c, dto.NewResellerProductSettingListResp(productSettingDTOInputList(rows)), response.BuildPagination(page, pageSize, total))
}

// GetProductSetting 获取当前用户的单个商品分销配置详情。
func (h *UserProductSettingHandler) GetProductSetting(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	productID, err := shared.ParseParamUint(c, "product_id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	detail, err := h.service.GetUserProductSetting(uid, productID)
	if err != nil {
		respondUserProductSettingError(c, err, "error.user_fetch_failed")
		return
	}
	response.Success(c, dto.NewResellerProductSettingDetailResp(productSettingDTOInputFromDetail(detail)))
}

// UpdateProductSettings 保存当前用户的商品级或 SKU 级分销配置。
func (h *UserProductSettingHandler) UpdateProductSettings(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	productID, err := shared.ParseParamUint(c, "product_id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	var req productSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	input, err := req.toInput()
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	detail, err := h.service.SaveUserProductSettings(uid, productID, input)
	if err != nil {
		respondUserProductSettingError(c, err, "error.save_failed")
		return
	}
	response.Success(c, dto.NewResellerProductSettingDetailResp(productSettingDTOInputFromDetail(detail)))
}

// PreviewProductSettings 计算当前用户拟用定价规则的预计生效价与校验结果（不落库）。
func (h *UserProductSettingHandler) PreviewProductSettings(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	productID, err := shared.ParseParamUint(c, "product_id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	var req productSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	input, err := req.toInput()
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	items, err := h.service.PreviewUserProductSettings(uid, productID, input)
	if err != nil {
		respondUserProductSettingError(c, err, "error.user_fetch_failed")
		return
	}
	previews := make([]dto.ResellerProductSettingPreviewInput, 0, len(items))
	for _, item := range items {
		previews = append(previews, dto.ResellerProductSettingPreviewInput{
			SKUID:          item.SKUID,
			IsListed:       item.IsListed,
			BasePrice:      item.BasePrice.StringFixed(2),
			EffectivePrice: item.EffectivePrice.StringFixed(2),
			Valid:          item.Valid,
			ErrorCode:      item.ErrorCode,
		})
	}
	response.Success(c, dto.NewResellerProductSettingPreviewResp(previews))
}

// ResetProductSetting 删除当前用户的商品级或 SKU 级分销配置。
func (h *UserProductSettingHandler) ResetProductSetting(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	productID, err := shared.ParseParamUint(c, "product_id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	skuID, err := shared.ParseQueryUint(c.Query("sku_id"), false)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	if err := h.service.ResetUserProductSetting(uid, productID, skuID); err != nil {
		respondUserProductSettingError(c, err, "error.save_failed")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func productSettingDTOInputFromDetail(detail *resellermodule.ProductSettingDetail) dto.ResellerProductSettingDTOInput {
	if detail == nil {
		return dto.ResellerProductSettingDTOInput{}
	}
	return dto.ResellerProductSettingDTOInput{
		Product:          detail.Product,
		Settings:         detail.Settings,
		EffectiveBySKUID: productSettingDecimalMapToStringMap(detail.EffectiveBySKUID),
		RuleBySKUID:      detail.RuleBySKUID,
	}
}

func productSettingDTOInputList(rows []resellermodule.ProductSettingListRow) []dto.ResellerProductSettingDTOInput {
	out := make([]dto.ResellerProductSettingDTOInput, 0, len(rows))
	for i := range rows {
		out = append(out, dto.ResellerProductSettingDTOInput{
			Product:          rows[i].Product,
			Settings:         rows[i].Settings,
			EffectiveBySKUID: productSettingDecimalMapToStringMap(rows[i].EffectiveBySKUID),
			RuleBySKUID:      rows[i].RuleBySKUID,
		})
	}
	return out
}

func productSettingDecimalMapToStringMap(input map[uint]decimal.Decimal) map[uint]string {
	out := make(map[uint]string, len(input))
	for key, value := range input {
		out[key] = value.StringFixed(2)
	}
	return out
}
