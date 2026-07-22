package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/dujiao-next/internal/repository"
)

const (
	walletDefaultCurrency = "CNY"
)

// WalletService 钱包服务
type WalletService struct {
	walletRepo            repository.WalletRepository
	orderRepo             repository.OrderRepository
	refundRecordRepo      repository.OrderRefundRecordRepository
	userRepo              repository.UserRepository
	affiliateSvc          *AffiliateService
	settingService        *SettingService
	resellerAccountingSvc *ResellerAccountingService
}

// NewWalletService 创建钱包服务
func NewWalletService(
	walletRepo repository.WalletRepository,
	orderRepo repository.OrderRepository,
	refundRecordRepo repository.OrderRefundRecordRepository,
	userRepo repository.UserRepository,
	affiliateSvc *AffiliateService,
	settingService *SettingService,
) *WalletService {
	return &WalletService{
		walletRepo:       walletRepo,
		orderRepo:        orderRepo,
		refundRecordRepo: refundRecordRepo,
		userRepo:         userRepo,
		affiliateSvc:     affiliateSvc,
		settingService:   settingService,
	}
}

func (s *WalletService) SetResellerAccountingService(svc *ResellerAccountingService) {
	s.resellerAccountingSvc = svc
}

func normalizeWalletCurrency(currency string) string {
	normalized := strings.ToUpper(strings.TrimSpace(currency))
	if normalized == "" {
		return walletDefaultCurrency
	}
	return normalized
}

func cleanWalletRemark(raw string, fallback string) string {
	remark := strings.TrimSpace(raw)
	if remark == "" {
		return fallback
	}
	return remark
}

func buildOrderWalletReference(orderID uint, action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		action = "wallet"
	}
	return fmt.Sprintf("order:%d:%s", orderID, action)
}

func buildWalletReference(prefix string, id uint) string {
	normalized := strings.TrimSpace(prefix)
	if normalized == "" {
		normalized = "wallet"
	}
	return fmt.Sprintf("%s:%d:%d", normalized, id, time.Now().UnixNano())
}
