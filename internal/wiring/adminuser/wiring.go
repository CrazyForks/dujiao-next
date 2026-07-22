package adminuserwiring

import (
	"github.com/dujiao-next/internal/provider"
	adminusertransport "github.com/dujiao-next/internal/transport/http/adminuser"
)

func NewHandler(c *provider.Container) *adminusertransport.AdminHandler {
	return adminusertransport.NewAdminHandler(
		adminUserDirectoryAdapter{users: c.UserRepo},
		adminUserEmailAdapter{},
		adminUserWalletAdapter{wallets: c.WalletService},
		adminUserOAuthAdapter{identities: c.ExternalIdentityStore},
		adminUserTelegramAdapter{auth: c.UserAuthService},
		adminUserCouponUsageAdapter{usages: c.CouponUsageRepo},
		adminUserCouponAdapter{coupons: c.CouponRepo},
		adminUserProductAdapter{products: c.ProductRepo},
		adminUserAuthStateAdapter{},
	)
}
