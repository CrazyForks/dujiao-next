package presenter

import (
	"time"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/models"
	giftcarddomain "github.com/dujiao-next/internal/modules/giftcard/domain"
	"github.com/dujiao-next/internal/shared/money"
)

// GiftCardRedeemResp 礼品卡兑换结果响应
type GiftCardRedeemResp struct {
	GiftCard    GiftCardResp          `json:"gift_card"`
	Wallet      dto.WalletAccountResp     `json:"wallet"`
	Transaction dto.WalletTransactionResp `json:"transaction"`
	WalletDelta money.Amount          `json:"wallet_delta"`
}

// GiftCardResp 礼品卡响应（兑换后）
type GiftCardResp struct {
	ID         uint         `json:"id"`
	Name       string       `json:"name"`
	Code       string       `json:"code"`
	Amount     money.Amount `json:"amount"`
	Currency   string       `json:"currency"`
	Status     string       `json:"status"`
	RedeemedAt *time.Time   `json:"redeemed_at"`
}

// NewGiftCardResp 从 giftcarddomain.GiftCard 构造响应
func NewGiftCardResp(c *giftcarddomain.GiftCard) GiftCardResp {
	return GiftCardResp{
		ID:         c.ID,
		Name:       c.Name,
		Code:       c.Code,
		Amount:     c.Amount,
		Currency:   c.Currency,
		Status:     c.Status,
		RedeemedAt: c.RedeemedAt,
	}
	// 排除：BatchID、ExpiresAt、RedeemedUserID、WalletTxnID、CreatedAt、UpdatedAt、Batch
}

// NewGiftCardRedeemResp 构造完整兑换响应
func NewGiftCardRedeemResp(card *giftcarddomain.GiftCard, account *models.WalletAccount, txn *models.WalletTransaction) GiftCardRedeemResp {
	return GiftCardRedeemResp{
		GiftCard:    NewGiftCardResp(card),
		Wallet:      dto.NewWalletAccountResp(account),
		Transaction: dto.NewWalletTransactionResp(txn),
		WalletDelta: card.Amount,
	}
}
