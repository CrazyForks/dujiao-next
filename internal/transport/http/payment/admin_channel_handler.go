package paymenthttp

import (
	"errors"

	ginutil "github.com/dujiao-next/internal/platform/http/ginutil"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/gin-gonic/gin"
)

// AdminChannelListFilter 后台支付渠道列表过滤条件。
type AdminChannelListFilter struct {
	Page         int
	PageSize     int
	ProviderType string
	ChannelType  string
	ActiveOnly   bool
}

// AdminChannelCatalog 后台支付渠道管理端口。
type AdminChannelCatalog interface {
	ValidateChannel(channel *models.PaymentChannel) error
	GetChannel(id uint) (*models.PaymentChannel, error)
	ListChannels(filter AdminChannelListFilter) ([]models.PaymentChannel, int64, error)
	Create(channel *models.PaymentChannel) error
	Update(channel *models.PaymentChannel) error
	Delete(id uint) error
}

// CreatePaymentChannelRequest 创建支付渠道请求
type CreatePaymentChannelRequest struct {
	Name               string                 `json:"name" binding:"required"`
	Icon               *string                `json:"icon"`
	ProviderType       string                 `json:"provider_type" binding:"required"`
	ChannelType        string                 `json:"channel_type" binding:"required"`
	InteractionMode    string                 `json:"interaction_mode" binding:"required"`
	FeeRate            *money.Amount          `json:"fee_rate"`
	FixedFee           *money.Amount          `json:"fixed_fee"`
	MinAmount          *money.Amount          `json:"min_amount"`
	MaxAmount          *money.Amount          `json:"max_amount"`
	HideAmountOutRange *bool                  `json:"hide_amount_out_range"`
	PaymentRoles       []string               `json:"payment_roles"`
	MemberLevels       []uint                 `json:"member_levels"`
	PaymentTypes       []string               `json:"payment_types"`
	ConfigJSON         map[string]interface{} `json:"config_json"`
	IsActive           *bool                  `json:"is_active"`
	SortOrder          int                    `json:"sort_order"`
}

// UpdatePaymentChannelRequest 更新支付渠道请求
type UpdatePaymentChannelRequest struct {
	Name               string                 `json:"name"`
	Icon               *string                `json:"icon"`
	ProviderType       string                 `json:"provider_type"`
	ChannelType        string                 `json:"channel_type"`
	InteractionMode    string                 `json:"interaction_mode"`
	FeeRate            *money.Amount          `json:"fee_rate"`
	FixedFee           *money.Amount          `json:"fixed_fee"`
	MinAmount          *money.Amount          `json:"min_amount"`
	MaxAmount          *money.Amount          `json:"max_amount"`
	HideAmountOutRange *bool                  `json:"hide_amount_out_range"`
	PaymentRoles       []string               `json:"payment_roles"`
	MemberLevels       []uint                 `json:"member_levels"`
	PaymentTypes       []string               `json:"payment_types"`
	ConfigJSON         map[string]interface{} `json:"config_json"`
	IsActive           *bool                  `json:"is_active"`
	SortOrder          *int                   `json:"sort_order"`
}

// AdminChannelHandler 处理后台支付渠道 HTTP。
type AdminChannelHandler struct {
	channels AdminChannelCatalog
}

func NewAdminChannelHandler(channels AdminChannelCatalog) *AdminChannelHandler {
	if channels == nil {
		panic("payment admin channel handler: channels is nil")
	}
	return &AdminChannelHandler{channels: channels}
}

// CreatePaymentChannel 创建支付渠道
func (h *AdminChannelHandler) CreatePaymentChannel(c *gin.Context) {
	var req CreatePaymentChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}

	channel := &models.PaymentChannel{
		Name:            req.Name,
		ProviderType:    req.ProviderType,
		ChannelType:     req.ChannelType,
		InteractionMode: req.InteractionMode,
		ConfigJSON:      jsonmap.JSON(req.ConfigJSON),
		PaymentRoles:    req.PaymentRoles,
		MemberLevels:    req.MemberLevels,
		PaymentTypes:    req.PaymentTypes,
		SortOrder:       req.SortOrder,
		IsActive:        true,
	}
	if req.Icon != nil {
		channel.Icon = *req.Icon
	}
	if req.IsActive != nil {
		channel.IsActive = *req.IsActive
	}
	if req.HideAmountOutRange != nil {
		channel.HideAmountOutRange = *req.HideAmountOutRange
	}
	if req.FeeRate != nil {
		channel.FeeRate = *req.FeeRate
	}
	if req.FixedFee != nil {
		channel.FixedFee = *req.FixedFee
	}
	if req.MinAmount != nil {
		channel.MinAmount = *req.MinAmount
	}
	if req.MaxAmount != nil {
		channel.MaxAmount = *req.MaxAmount
	}

	if err := h.channels.ValidateChannel(channel); err != nil {
		switch {
		case errors.Is(err, ErrPaymentProviderNotSupported):
			ginutil.RespondError(c, response.CodeBadRequest, "error.payment_provider_not_supported", nil)
		case errors.Is(err, ErrPaymentChannelConfigInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.payment_channel_config_invalid", nil)
		default:
			ginutil.RespondError(c, response.CodeBadRequest, "error.payment_channel_invalid", nil)
		}
		return
	}

	if err := h.channels.Create(channel); err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.payment_channel_create_failed", err)
		return
	}
	_ = cache.DelAllPublicConfig(c.Request.Context())

	response.Success(c, channel)
}

// UpdatePaymentChannel 更新支付渠道
func (h *AdminChannelHandler) UpdatePaymentChannel(c *gin.Context) {
	id, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.payment_channel_invalid", nil)
		return
	}

	var req UpdatePaymentChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}

	channel, err := h.channels.GetChannel(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentChannelNotFound):
			ginutil.RespondError(c, response.CodeNotFound, "error.payment_channel_not_found", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.payment_channel_update_failed", err)
		}
		return
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Icon != nil {
		channel.Icon = *req.Icon
	}
	if req.ProviderType != "" {
		channel.ProviderType = req.ProviderType
	}
	if req.ChannelType != "" {
		channel.ChannelType = req.ChannelType
	}
	if req.InteractionMode != "" {
		channel.InteractionMode = req.InteractionMode
	}
	if req.FeeRate != nil {
		channel.FeeRate = *req.FeeRate
	}
	if req.FixedFee != nil {
		channel.FixedFee = *req.FixedFee
	}
	if req.MinAmount != nil {
		channel.MinAmount = *req.MinAmount
	}
	if req.MaxAmount != nil {
		channel.MaxAmount = *req.MaxAmount
	}
	if req.PaymentRoles != nil {
		channel.PaymentRoles = req.PaymentRoles
	}
	if req.MemberLevels != nil {
		channel.MemberLevels = req.MemberLevels
	}
	if req.PaymentTypes != nil {
		channel.PaymentTypes = req.PaymentTypes
	}
	if req.ConfigJSON != nil {
		channel.ConfigJSON = jsonmap.JSON(req.ConfigJSON)
	}
	if req.IsActive != nil {
		channel.IsActive = *req.IsActive
	}
	if req.HideAmountOutRange != nil {
		channel.HideAmountOutRange = *req.HideAmountOutRange
	}
	if req.SortOrder != nil {
		channel.SortOrder = *req.SortOrder
	}

	if err := h.channels.ValidateChannel(channel); err != nil {
		switch {
		case errors.Is(err, ErrPaymentProviderNotSupported):
			ginutil.RespondError(c, response.CodeBadRequest, "error.payment_provider_not_supported", nil)
		case errors.Is(err, ErrPaymentChannelConfigInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.payment_channel_config_invalid", nil)
		default:
			ginutil.RespondError(c, response.CodeBadRequest, "error.payment_channel_invalid", nil)
		}
		return
	}

	if err := h.channels.Update(channel); err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.payment_channel_update_failed", err)
		return
	}
	_ = cache.DelAllPublicConfig(c.Request.Context())

	response.Success(c, channel)
}

// DeletePaymentChannel 删除支付渠道
func (h *AdminChannelHandler) DeletePaymentChannel(c *gin.Context) {
	id, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.payment_channel_invalid", nil)
		return
	}

	if err := h.channels.Delete(id); err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.payment_channel_delete_failed", err)
		return
	}
	_ = cache.DelAllPublicConfig(c.Request.Context())

	response.Success(c, gin.H{"deleted": true})
}

// GetPaymentChannel 获取支付渠道详情
func (h *AdminChannelHandler) GetPaymentChannel(c *gin.Context) {
	id, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.payment_channel_invalid", nil)
		return
	}

	channel, err := h.channels.GetChannel(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentChannelNotFound):
			ginutil.RespondError(c, response.CodeNotFound, "error.payment_channel_not_found", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.payment_channel_fetch_failed", err)
		}
		return
	}

	response.Success(c, channel)
}

// GetPaymentChannels 获取支付渠道列表
func (h *AdminChannelHandler) GetPaymentChannels(c *gin.Context) {
	page, pageSize := ginutil.ParsePagination(c)

	providerType := c.Query("provider_type")
	channelType := c.Query("channel_type")
	activeOnly, err := ginutil.ParseQueryBool(c, "active_only")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}

	channels, total, err := h.channels.ListChannels(AdminChannelListFilter{
		Page:         page,
		PageSize:     pageSize,
		ProviderType: providerType,
		ChannelType:  channelType,
		ActiveOnly:   activeOnly,
	})
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.payment_channel_fetch_failed", err)
		return
	}

	pagination := response.BuildPagination(page, pageSize, total)
	response.SuccessWithPage(c, channels, pagination)
}
