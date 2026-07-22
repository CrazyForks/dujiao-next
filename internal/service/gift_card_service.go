package service

import (
	"fmt"
	"strings"
	"time"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/giftcard"
	giftcardgormstore "github.com/dujiao-next/internal/modules/giftcard/store/gormstore"
	"github.com/dujiao-next/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GiftCardService 礼品卡服务（管理用例委托 modules；兑换仍留在此薄壳）。
type GiftCardService struct {
	admin         *giftcard.Service
	store         *giftcardgormstore.Store
	walletService *WalletService
}

type GenerateGiftCardsInput = giftcard.GenerateInput
type GiftCardListInput = giftcard.ListInput
type UpdateGiftCardInput = giftcard.UpdateInput
type GiftCardRedeemInput = giftcard.RedeemInput

// NewGiftCardService 创建礼品卡服务
func NewGiftCardService(store *giftcardgormstore.Store, userRepo repository.UserRepository, walletService *WalletService, settingSvc *settingsapp.Service) *GiftCardService {
	return &GiftCardService{
		admin:         newGiftCardAdminService(store, userRepo, settingSvc),
		store:         store,
		walletService: walletService,
	}
}

// GenerateGiftCards 生成礼品卡批次
func (s *GiftCardService) GenerateGiftCards(input GenerateGiftCardsInput) (*models.GiftCardBatch, int, error) {
	if s == nil || s.admin == nil {
		return nil, 0, ErrGiftCardCreateFailed
	}
	return s.admin.Generate(input)
}

// ListGiftCards 获取礼品卡列表
func (s *GiftCardService) ListGiftCards(input GiftCardListInput) ([]models.GiftCard, int64, error) {
	if s == nil || s.admin == nil {
		return nil, 0, ErrGiftCardFetchFailed
	}
	return s.admin.List(input)
}

// UpdateGiftCard 更新礼品卡
func (s *GiftCardService) UpdateGiftCard(id uint, input UpdateGiftCardInput) (*models.GiftCard, error) {
	if s == nil || s.admin == nil {
		return nil, ErrGiftCardInvalid
	}
	return s.admin.Update(id, input)
}

// DeleteGiftCard 删除礼品卡
func (s *GiftCardService) DeleteGiftCard(id uint) error {
	if s == nil || s.admin == nil {
		return ErrGiftCardInvalid
	}
	return s.admin.Delete(id)
}

// BatchUpdateStatus 批量更新礼品卡状态
func (s *GiftCardService) BatchUpdateStatus(ids []uint, status string) (int64, error) {
	if s == nil || s.admin == nil {
		return 0, ErrGiftCardInvalid
	}
	return s.admin.BatchUpdateStatus(ids, status)
}

// ExportGiftCards 导出礼品卡
func (s *GiftCardService) ExportGiftCards(ids []uint, format string) ([]byte, string, error) {
	if s == nil || s.admin == nil {
		return nil, "", ErrGiftCardFetchFailed
	}
	return s.admin.Export(ids, format)
}

// ResolveRedeemedUsers 批量解析礼品卡兑换用户
func (s *GiftCardService) ResolveRedeemedUsers(cards []models.GiftCard) (map[uint]models.User, error) {
	if s == nil || s.admin == nil {
		return map[uint]models.User{}, nil
	}
	return s.admin.ResolveRedeemedUsers(cards)
}

// RedeemGiftCard 兑换礼品卡
func (s *GiftCardService) RedeemGiftCard(input GiftCardRedeemInput) (*models.GiftCard, *models.WalletAccount, *models.WalletTransaction, error) {
	if s == nil || s.store == nil || s.walletService == nil {
		return nil, nil, nil, ErrGiftCardFetchFailed
	}
	code := strings.TrimSpace(strings.ToUpper(input.Code))
	if input.UserID == 0 || code == "" {
		return nil, nil, nil, ErrGiftCardInvalid
	}

	var (
		resultCard  *models.GiftCard
		resultAcc   *models.WalletAccount
		resultTxn   *models.WalletTransaction
		resultError error
	)
	err := s.store.Transaction(func(tx *gorm.DB) error {
		repo := s.store.WithTx(tx)
		card, err := repo.GetByCodeForUpdate(code)
		if err != nil {
			return ErrGiftCardFetchFailed
		}
		if card == nil {
			return ErrGiftCardNotFound
		}
		switch card.Status {
		case models.GiftCardStatusRedeemed:
			return ErrGiftCardRedeemed
		case models.GiftCardStatusDisabled:
			return ErrGiftCardDisabled
		case models.GiftCardStatusActive:
		default:
			return ErrGiftCardInvalid
		}
		if isGiftCardExpired(card.ExpiresAt, time.Now()) {
			return ErrGiftCardExpired
		}
		if card.Amount.Decimal.Round(2).LessThanOrEqual(decimal.Zero) {
			return ErrGiftCardInvalid
		}

		now := time.Now()
		account, txn, creditErr := s.walletService.CreditInTx(tx, WalletCreditInput{
			UserID:    input.UserID,
			Amount:    card.Amount,
			Currency:  card.Currency,
			TxnType:   constants.WalletTxnTypeGiftCard,
			Reference: fmt.Sprintf("gift_card:%d", card.ID),
			Remark:    fmt.Sprintf("礼品卡兑换：%s", card.Code),
			OrderID:   nil,
		})
		if creditErr != nil {
			return creditErr
		}

		card.Status = models.GiftCardStatusRedeemed
		card.RedeemedUserID = &input.UserID
		card.RedeemedAt = &now
		if txn != nil && txn.ID > 0 {
			card.WalletTxnID = &txn.ID
		}
		card.UpdatedAt = now
		if err := repo.Update(card); err != nil {
			return ErrGiftCardUpdateFailed
		}
		resultCard = card
		resultAcc = account
		resultTxn = txn
		return nil
	})
	if err != nil {
		resultError = err
	}
	return resultCard, resultAcc, resultTxn, resultError
}

func isGiftCardExpired(expiresAt *time.Time, now time.Time) bool {
	if expiresAt == nil || expiresAt.IsZero() {
		return false
	}
	return expiresAt.Before(now)
}
