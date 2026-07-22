package giftcardhttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/modules/giftcard"

	"github.com/gin-gonic/gin"
)

// ChannelUserProvisioner 是渠道礼品卡兑换所需的身份开通端口。
type ChannelUserProvisioner interface {
	ProvisionUserID(channelUserID string) (uint, error)
}

// ChannelHandler 处理渠道礼品卡请求。
type ChannelHandler struct {
	cards UserService
	users ChannelUserProvisioner
}

func NewChannelHandler(cards UserService, users ChannelUserProvisioner) *ChannelHandler {
	if cards == nil {
		panic("giftcard channel handler: cards is nil")
	}
	if users == nil {
		panic("giftcard channel handler: users is nil")
	}
	return &ChannelHandler{cards: cards, users: users}
}

type channelRedeemRequest struct {
	ChannelUserID  string `json:"channel_user_id"`
	TelegramUserID string `json:"telegram_user_id"`
	Code           string `json:"code" binding:"required"`
}

// Redeem POST /api/v1/channel/wallet/gift-card/redeem
func (h *ChannelHandler) Redeem(c *gin.Context) {
	var req channelRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.ChannelBindError(c, err)
		return
	}

	channelUserID := shared.ChannelUserIDValue(req.ChannelUserID, req.TelegramUserID)
	if channelUserID == "" {
		shared.ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.bad_request", nil)
		return
	}

	userID, err := h.users.ProvisionUserID(channelUserID)
	if err != nil {
		shared.RequestLog(c).Errorw("channel_wallet_gift_card_resolve_user", "channel_user_id", channelUserID, "error", err)
		shared.ChannelIdentityError(c, err)
		return
	}

	card, account, txn, err := h.cards.RedeemGiftCard(giftcard.RedeemInput{
		UserID: userID,
		Code:   strings.TrimSpace(req.Code),
	})
	if err != nil {
		shared.RequestLog(c).Warnw("channel_wallet_gift_card_redeem_failed", "user_id", userID, "channel_user_id", channelUserID, "error", err)
		respondChannelGiftCardError(c, err)
		return
	}

	shared.ChannelSuccess(c, dto.NewGiftCardRedeemResp(card, account, txn))
}

func respondChannelGiftCardError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, giftcard.ErrInvalid):
		shared.ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "gift_card_invalid", "error.gift_card_invalid", nil)
	case errors.Is(err, giftcard.ErrNotFound):
		shared.ChannelError(c, http.StatusNotFound, response.CodeNotFound, "gift_card_not_found", "error.gift_card_not_found", nil)
	case errors.Is(err, giftcard.ErrExpired):
		shared.ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "gift_card_expired", "error.gift_card_expired", nil)
	case errors.Is(err, giftcard.ErrDisabled):
		shared.ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "gift_card_disabled", "error.gift_card_disabled", nil)
	case errors.Is(err, giftcard.ErrRedeemed):
		shared.ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "gift_card_redeemed", "error.gift_card_redeemed", nil)
	default:
		shared.ChannelError(c, http.StatusInternalServerError, response.CodeInternal, "gift_card_redeem_failed", "error.gift_card_redeem_failed", err)
	}
}
