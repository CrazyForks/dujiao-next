package affiliatewiring

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/affiliate"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
)

// affiliateAdminAdapter 将 legacy AffiliateService 适配为后台 transport 端口。
type affiliateAdminAdapter struct {
	svc *service.AffiliateService
}

func (a affiliateAdminAdapter) ListAdminUsers(filter affiliate.AdminProfileListFilter) ([]affiliate.AdminUserItem, int64, error) {
	return a.svc.ListAdminUsers(repository.AffiliateProfileListFilter{
		Page:     filter.Page,
		PageSize: filter.PageSize,
		UserID:   filter.UserID,
		Status:   filter.Status,
		Code:     filter.Code,
		Keyword:  filter.Keyword,
	})
}

func (a affiliateAdminAdapter) ListAdminCommissions(filter affiliate.AdminCommissionListFilter) ([]models.AffiliateCommission, int64, error) {
	return a.svc.ListAdminCommissions(filter)
}

func (a affiliateAdminAdapter) ListAdminWithdraws(filter affiliate.AdminWithdrawListFilter) ([]models.AffiliateWithdrawRequest, int64, error) {
	return a.svc.ListAdminWithdraws(filter)
}

func (a affiliateAdminAdapter) UpdateAffiliateProfileStatus(profileID uint, status string) (*models.AffiliateProfile, error) {
	return a.svc.UpdateAffiliateProfileStatus(profileID, status)
}

func (a affiliateAdminAdapter) BatchUpdateAffiliateProfileStatus(profileIDs []uint, status string) (int64, error) {
	return a.svc.BatchUpdateAffiliateProfileStatus(profileIDs, status)
}

func (a affiliateAdminAdapter) ReviewWithdraw(adminID, withdrawID uint, action, reason string) (*models.AffiliateWithdrawRequest, error) {
	return a.svc.ReviewWithdraw(adminID, withdrawID, action, reason)
}
