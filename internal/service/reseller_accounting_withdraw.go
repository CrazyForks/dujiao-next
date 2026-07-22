package service

import (
	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
)

type ResellerWithdrawApplyInput = resellermodule.WithdrawApplyInput

func (s *ResellerAccountingService) ApplyUserWithdraw(userID uint, input ResellerWithdrawApplyInput) (*models.ResellerWithdrawRequest, error) {
	if s == nil || s.withdraw == nil {
		return nil, ErrResellerAccountingUnavailable
	}
	return s.withdraw.ApplyUserWithdraw(userID, input)
}

func (s *ResellerAccountingService) ApplyWithdraw(resellerID uint, input ResellerWithdrawApplyInput) (*models.ResellerWithdrawRequest, error) {
	if s == nil || s.withdraw == nil {
		return nil, ErrResellerAccountingUnavailable
	}
	return s.withdraw.ApplyWithdraw(resellerID, input)
}

func (s *ResellerAccountingService) ReviewWithdraw(adminID uint, withdrawID uint, action string, rejectReason string) (*models.ResellerWithdrawRequest, error) {
	if s == nil || s.withdraw == nil {
		return nil, ErrNotFound
	}
	return s.withdraw.ReviewWithdraw(adminID, withdrawID, action, rejectReason)
}
