package service

import (
	"math"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/affiliate"
	"github.com/dujiao-next/internal/repository"

	"github.com/shopspring/decimal"
)

// 兼容门面别名。
type (
	AffiliateDashboard                 = affiliate.Dashboard
	AffiliateStats                     = affiliate.Stats
	AffiliateAdminUserItem             = affiliate.AdminUserItem
	AffiliateAdminCommissionListFilter = affiliate.AdminCommissionListFilter
	AffiliateAdminWithdrawListFilter   = affiliate.AdminWithdrawListFilter
)

// GetUserDashboard 获取用户返利中心数据
func (s *AffiliateService) GetUserDashboard(userID uint) (AffiliateDashboard, error) {
	dashboard := AffiliateDashboard{
		Opened:              false,
		PendingCommission:   models.NewMoneyFromDecimal(decimal.Zero),
		AvailableCommission: models.NewMoneyFromDecimal(decimal.Zero),
		WithdrawnCommission: models.NewMoneyFromDecimal(decimal.Zero),
	}
	if userID == 0 || s.repo == nil {
		return dashboard, nil
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return dashboard, err
	}
	if profile == nil {
		return dashboard, nil
	}

	stats, err := s.buildProfileStats(profile.ID)
	if err != nil {
		return dashboard, err
	}
	dashboard.Opened = true
	dashboard.AffiliateCode = profile.AffiliateCode
	dashboard.PromotionPath = "/?aff=" + profile.AffiliateCode
	dashboard.ClickCount = stats.ClickCount
	dashboard.ValidOrderCount = stats.ValidOrderCount
	dashboard.ConversionRate = stats.ConversionRate
	dashboard.PendingCommission = stats.PendingCommission
	dashboard.AvailableCommission = stats.AvailableCommission
	dashboard.WithdrawnCommission = stats.WithdrawnCommission
	return dashboard, nil
}

// ListUserCommissions 查询用户佣金记录
func (s *AffiliateService) ListUserCommissions(userID uint, page, pageSize int, status string) ([]models.AffiliateCommission, int64, error) {
	if userID == 0 || s.repo == nil {
		return []models.AffiliateCommission{}, 0, nil
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if profile == nil {
		return []models.AffiliateCommission{}, 0, nil
	}
	return s.repo.ListCommissions(repository.AffiliateCommissionListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profile.ID,
		Status:             strings.TrimSpace(status),
	})
}

// ListUserWithdraws 查询用户提现记录
func (s *AffiliateService) ListUserWithdraws(userID uint, page, pageSize int, status string) ([]models.AffiliateWithdrawRequest, int64, error) {
	if userID == 0 || s.repo == nil {
		return []models.AffiliateWithdrawRequest{}, 0, nil
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if profile == nil {
		return []models.AffiliateWithdrawRequest{}, 0, nil
	}
	return s.repo.ListWithdraws(repository.AffiliateWithdrawListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profile.ID,
		Status:             strings.TrimSpace(status),
	})
}

// ListAdminUsers 后台查询推广用户列表
func (s *AffiliateService) ListAdminUsers(filter repository.AffiliateProfileListFilter) ([]AffiliateAdminUserItem, int64, error) {
	if s.repo == nil {
		return []AffiliateAdminUserItem{}, 0, nil
	}
	rows, total, err := s.repo.ListProfiles(filter)
	if err != nil {
		return nil, 0, err
	}
	profileIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.ID == 0 {
			continue
		}
		profileIDs = append(profileIDs, row.ID)
	}
	statsMap, err := s.repo.GetProfileStatsBatch(profileIDs)
	if err != nil {
		return nil, 0, err
	}
	result := make([]AffiliateAdminUserItem, 0, len(rows))
	for _, row := range rows {
		agg := statsMap[row.ID]
		stats := AffiliateStats{
			ClickCount:          agg.ClickCount,
			ValidOrderCount:     agg.ValidOrderCount,
			ConversionRate:      calcAffiliateConversion(agg.ValidOrderCount, agg.ClickCount),
			PendingCommission:   models.NewMoneyFromDecimal(agg.PendingCommission.Round(2)),
			AvailableCommission: models.NewMoneyFromDecimal(agg.AvailableCommission.Round(2)),
			WithdrawnCommission: models.NewMoneyFromDecimal(agg.WithdrawnCommission.Round(2)),
		}
		result = append(result, AffiliateAdminUserItem{
			Profile: row,
			Stats:   stats,
		})
	}
	return result, total, nil
}

// ListAdminCommissions 后台查询佣金记录
func (s *AffiliateService) ListAdminCommissions(filter AffiliateAdminCommissionListFilter) ([]models.AffiliateCommission, int64, error) {
	if s.repo == nil {
		return []models.AffiliateCommission{}, 0, nil
	}
	return s.repo.ListCommissions(repository.AffiliateCommissionListFilter{
		Page:               filter.Page,
		PageSize:           filter.PageSize,
		AffiliateProfileID: filter.AffiliateProfileID,
		OrderNo:            strings.TrimSpace(filter.OrderNo),
		Status:             strings.TrimSpace(filter.Status),
		Keyword:            strings.TrimSpace(filter.Keyword),
	})
}

// ListAdminWithdraws 后台查询提现申请
func (s *AffiliateService) ListAdminWithdraws(filter AffiliateAdminWithdrawListFilter) ([]models.AffiliateWithdrawRequest, int64, error) {
	if s.repo == nil {
		return []models.AffiliateWithdrawRequest{}, 0, nil
	}
	return s.repo.ListWithdraws(repository.AffiliateWithdrawListFilter{
		Page:               filter.Page,
		PageSize:           filter.PageSize,
		AffiliateProfileID: filter.AffiliateProfileID,
		Status:             strings.TrimSpace(filter.Status),
		Keyword:            strings.TrimSpace(filter.Keyword),
	})
}

func (s *AffiliateService) buildProfileStats(profileID uint) (AffiliateStats, error) {
	stats := AffiliateStats{
		PendingCommission:   models.NewMoneyFromDecimal(decimal.Zero),
		AvailableCommission: models.NewMoneyFromDecimal(decimal.Zero),
		WithdrawnCommission: models.NewMoneyFromDecimal(decimal.Zero),
	}
	if profileID == 0 || s.repo == nil {
		return stats, nil
	}
	clickCount, err := s.repo.CountClicksByProfile(profileID)
	if err != nil {
		return stats, err
	}
	validOrders, err := s.repo.CountValidOrdersByProfile(profileID)
	if err != nil {
		return stats, err
	}
	pendingAmount, err := s.repo.SumCommissionByProfile(profileID, []string{
		constants.AffiliateCommissionStatusPendingConfirm,
	}, false)
	if err != nil {
		return stats, err
	}
	availableAmount, err := s.repo.SumCommissionByProfile(profileID, []string{
		constants.AffiliateCommissionStatusAvailable,
	}, true)
	if err != nil {
		return stats, err
	}
	withdrawnAmount, err := s.repo.SumCommissionByProfile(profileID, []string{
		constants.AffiliateCommissionStatusWithdrawn,
	}, false)
	if err != nil {
		return stats, err
	}

	stats.ClickCount = clickCount
	stats.ValidOrderCount = validOrders
	stats.ConversionRate = calcAffiliateConversion(validOrders, clickCount)
	stats.PendingCommission = models.NewMoneyFromDecimal(pendingAmount)
	stats.AvailableCommission = models.NewMoneyFromDecimal(availableAmount)
	stats.WithdrawnCommission = models.NewMoneyFromDecimal(withdrawnAmount)
	return stats, nil
}

func calcAffiliateConversion(validOrders, clicks int64) float64 {
	if clicks <= 0 || validOrders <= 0 {
		return 0
	}
	value := (float64(validOrders) / float64(clicks)) * 100
	return math.Round(value*100) / 100
}
