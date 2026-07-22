package giftcardhttp

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/modules/giftcard"

	"github.com/gin-gonic/gin"
)

// CaptchaVerifier 是礼品卡兑换所需的验证码端口。
type CaptchaVerifier interface {
	Verify(scene string, payload shared.CaptchaPayloadRequest, clientIP string) error
}

// UserService 是用户侧礼品卡兑换端口。
type UserService interface {
	RedeemGiftCard(input giftcard.RedeemInput) (*models.GiftCard, *models.WalletAccount, *models.WalletTransaction, error)
}

// UserHandler 处理用户中心礼品卡请求。
type UserHandler struct {
	cards   UserService
	captcha CaptchaVerifier
}

func NewUserHandler(cards UserService, captcha CaptchaVerifier) *UserHandler {
	if cards == nil {
		panic("giftcard user handler: cards is nil")
	}
	return &UserHandler{cards: cards, captcha: captcha}
}

type redeemRequest struct {
	Code           string                       `json:"code" binding:"required"`
	CaptchaPayload shared.CaptchaPayloadRequest `json:"captcha_payload"`
}

// Redeem 用户兑换礼品卡。
func (h *UserHandler) Redeem(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	var req redeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if h.captcha != nil {
		if captchaErr := h.captcha.Verify(constants.CaptchaSceneGiftCardRedeem, req.CaptchaPayload, c.ClientIP()); captchaErr != nil {
			respondCaptchaError(c, captchaErr)
			return
		}
	}
	card, account, txn, err := h.cards.RedeemGiftCard(giftcard.RedeemInput{
		UserID: uid,
		Code:   strings.TrimSpace(req.Code),
	})
	if err != nil {
		respondUserGiftCardError(c, err)
		return
	}
	response.Success(c, dto.NewGiftCardRedeemResp(card, account, txn))
}

func respondCaptchaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, captcha.ErrRequired):
		shared.RespondError(c, response.CodeBadRequest, "error.captcha_required", nil)
	case errors.Is(err, captcha.ErrInvalid):
		shared.RespondError(c, response.CodeBadRequest, "error.captcha_invalid", nil)
	case errors.Is(err, captcha.ErrConfigInvalid):
		shared.RespondError(c, response.CodeInternal, "error.captcha_config_invalid", err)
	default:
		shared.RespondError(c, response.CodeInternal, "error.captcha_verify_failed", err)
	}
}

func respondUserGiftCardError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, giftcard.ErrInvalid):
		shared.RespondError(c, response.CodeBadRequest, "error.gift_card_invalid", nil)
	case errors.Is(err, giftcard.ErrNotFound):
		shared.RespondError(c, response.CodeNotFound, "error.gift_card_not_found", nil)
	case errors.Is(err, giftcard.ErrExpired):
		shared.RespondError(c, response.CodeBadRequest, "error.gift_card_expired", nil)
	case errors.Is(err, giftcard.ErrDisabled):
		shared.RespondError(c, response.CodeBadRequest, "error.gift_card_disabled", nil)
	case errors.Is(err, giftcard.ErrRedeemed):
		shared.RespondError(c, response.CodeBadRequest, "error.gift_card_redeemed", nil)
	default:
		shared.RespondError(c, response.CodeInternal, "error.gift_card_redeem_failed", err)
	}
}
