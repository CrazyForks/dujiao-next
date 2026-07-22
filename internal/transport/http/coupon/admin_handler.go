package couponhttp

import (
	"errors"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/coupon"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AdminService interface {
	Create(input coupon.CreateCouponInput) (*models.Coupon, error)
	Update(id uint, input coupon.UpdateCouponInput) (*models.Coupon, error)
	Delete(id uint) error
	List(filter coupon.ListFilter) ([]models.Coupon, int64, error)
}

type AdminHandler struct {
	service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

// CreateCouponRequest 创建优惠券请求
type CreateCouponRequest struct {
	Code                   string   `json:"code" binding:"required"`
	Type                   string   `json:"type" binding:"required"`
	Value                  float64  `json:"value" binding:"required"`
	MinAmount              float64  `json:"min_amount"`
	MaxDiscount            float64  `json:"max_discount"`
	UsageLimit             int      `json:"usage_limit"`
	PerUserLimit           int      `json:"per_user_limit"`
	DisabledWholesalePrice *bool    `json:"disabled_wholesale_price"`
	PerItemDiscount        *bool    `json:"per_item_discount"`
	PaymentRoles           []string `json:"payment_roles"`
	MemberLevels           []uint   `json:"member_levels"`
	ScopeRefIDs            []uint   `json:"scope_ref_ids" binding:"required"`
	StartsAt               string   `json:"starts_at"`
	EndsAt                 string   `json:"ends_at"`
	IsActive               *bool    `json:"is_active"`
}

func buildCreateCouponInputFromRequest(req CreateCouponRequest) (coupon.CreateCouponInput, error) {
	startsAt, err := ginutil.ParseTimeNullable(req.StartsAt)
	if err != nil {
		return coupon.CreateCouponInput{}, err
	}
	endsAt, err := ginutil.ParseTimeNullable(req.EndsAt)
	if err != nil {
		return coupon.CreateCouponInput{}, err
	}
	return coupon.CreateCouponInput{
		Code:                   req.Code,
		Type:                   req.Type,
		Value:                  models.NewMoneyFromDecimal(decimal.NewFromFloat(req.Value)),
		MinAmount:              models.NewMoneyFromDecimal(decimal.NewFromFloat(req.MinAmount)),
		MaxDiscount:            models.NewMoneyFromDecimal(decimal.NewFromFloat(req.MaxDiscount)),
		UsageLimit:             req.UsageLimit,
		PerUserLimit:           req.PerUserLimit,
		DisabledWholesalePrice: req.DisabledWholesalePrice,
		PerItemDiscount:        req.PerItemDiscount,
		PaymentRoles:           req.PaymentRoles,
		MemberLevels:           req.MemberLevels,
		ScopeRefIDs:            req.ScopeRefIDs,
		StartsAt:               startsAt,
		EndsAt:                 endsAt,
		IsActive:               req.IsActive,
	}, nil
}

func buildUpdateCouponInputFromRequest(req CreateCouponRequest) (coupon.UpdateCouponInput, error) {
	input, err := buildCreateCouponInputFromRequest(req)
	if err != nil {
		return coupon.UpdateCouponInput{}, err
	}
	return coupon.UpdateCouponInput(input), nil
}

// CreateCoupon 创建优惠券
func (h *AdminHandler) CreateCoupon(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}

	input, err := buildCreateCouponInputFromRequest(req)
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}

	created, err := h.service.Create(input)
	if err != nil {
		switch {
		case errors.Is(err, coupon.ErrInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.coupon_invalid", nil)
		case errors.Is(err, coupon.ErrScopeInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.coupon_scope_invalid", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.coupon_create_failed", err)
		}
		return
	}

	response.Success(c, created)
}

// UpdateCoupon 更新优惠券
func (h *AdminHandler) UpdateCoupon(c *gin.Context) {
	couponID, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}

	input, err := buildUpdateCouponInputFromRequest(req)
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}

	updated, err := h.service.Update(couponID, input)
	if err != nil {
		switch {
		case errors.Is(err, coupon.ErrNotFound):
			ginutil.RespondError(c, response.CodeNotFound, "error.coupon_not_found", nil)
		case errors.Is(err, coupon.ErrInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.coupon_invalid", nil)
		case errors.Is(err, coupon.ErrScopeInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.coupon_scope_invalid", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.coupon_update_failed", err)
		}
		return
	}

	response.Success(c, updated)
}

// DeleteCoupon 删除优惠券
func (h *AdminHandler) DeleteCoupon(c *gin.Context) {
	couponID, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	if err := h.service.Delete(couponID); err != nil {
		switch {
		case errors.Is(err, coupon.ErrNotFound):
			ginutil.RespondError(c, response.CodeNotFound, "error.coupon_not_found", nil)
		case errors.Is(err, coupon.ErrInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.coupon_invalid", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.coupon_delete_failed", err)
		}
		return
	}
	response.Success(c, gin.H{
		"deleted": true,
	})
}

// GetAdminCoupons 获取优惠券列表
func (h *AdminHandler) GetAdminCoupons(c *gin.Context) {
	page, pageSize := ginutil.ParsePagination(c)

	code := c.Query("code")
	id, err := ginutil.ParseQueryUint(c.Query("id"), true)
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	scopeRefID, err := ginutil.ParseQueryUint(c.Query("scope_ref_id"), true)
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	isActive, err := ginutil.ParseQueryBoolPtr(c, "is_active")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}

	coupons, total, err := h.service.List(coupon.ListFilter{
		ID:         id,
		Code:       code,
		ScopeRefID: scopeRefID,
		IsActive:   isActive,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.coupon_fetch_failed", err)
		return
	}

	pagination := response.BuildPagination(page, pageSize, total)
	response.SuccessWithPage(c, coupons, pagination)
}
