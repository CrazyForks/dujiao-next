package adminauthzwiring

import (
	"github.com/dujiao-next/internal/provider"
	adminauthztransport "github.com/dujiao-next/internal/transport/http/adminauthz"
)

func NewHandler(c *provider.Container) *adminauthztransport.AdminHandler {
	return adminauthztransport.NewAdminHandler(
		adminAuthzRolePolicyAdapter{svc: c.AuthzService},
		adminAuthzDirectoryAdapter{admins: c.AdminRepo},
		adminAuthzPasswordAdapter{auth: c.AuthService},
		adminAuthzAuthStateAdapter{},
		adminAuthzAuditAdapter{svc: c.AuthzAuditService},
	)
}
