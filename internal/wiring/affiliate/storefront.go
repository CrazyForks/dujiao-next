package affiliatewiring

import (
	"github.com/dujiao-next/internal/modules/affiliate"
	affiliatedomain "github.com/dujiao-next/internal/modules/affiliate/domain"
	"github.com/dujiao-next/internal/service"
)

// affiliateStorefrontAdapter 将 legacy AffiliateService 适配为 transport 端口。
type affiliateStorefrontAdapter struct {
	svc *service.AffiliateService
}

func (a affiliateStorefrontAdapter) TrackClick(input affiliate.TrackClickInput) error {
	return a.svc.TrackClick(input)
}

func (a affiliateStorefrontAdapter) OpenAffiliate(userID uint) (*affiliatedomain.Profile, error) {
	return a.svc.OpenAffiliate(userID)
}

func (a affiliateStorefrontAdapter) GetUserDashboard(userID uint) (affiliate.Dashboard, error) {
	return a.svc.GetUserDashboard(userID)
}

func (a affiliateStorefrontAdapter) ListUserCommissions(userID uint, page, pageSize int, status string) ([]affiliatedomain.Commission, int64, error) {
	return a.svc.ListUserCommissions(userID, page, pageSize, status)
}

func (a affiliateStorefrontAdapter) ListUserWithdraws(userID uint, page, pageSize int, status string) ([]affiliatedomain.WithdrawRequest, int64, error) {
	return a.svc.ListUserWithdraws(userID, page, pageSize, status)
}

func (a affiliateStorefrontAdapter) ApplyWithdraw(userID uint, input affiliate.WithdrawApplyInput) (*affiliatedomain.WithdrawRequest, error) {
	return a.svc.ApplyWithdraw(userID, input)
}
