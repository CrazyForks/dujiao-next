package resellerwiring

import (
	"github.com/dujiao-next/internal/provider"
	resellertransport "github.com/dujiao-next/internal/transport/http/reseller"
)

type Handlers struct {
	User                *resellertransport.UserHandler
	UserProductSetting  *resellertransport.UserProductSettingHandler
	UserFinance         *resellertransport.UserFinanceHandler
	UserOrder           *resellertransport.UserOrderHandler
	AdminManagement     *resellertransport.AdminManagementHandler
	AdminProfileDetail  *resellertransport.AdminProfileDetailHandler
	AdminSiteConfig     *resellertransport.AdminSiteConfigHandler
	AdminProductSetting *resellertransport.AdminProductSettingHandler
	AdminOperations     *resellertransport.AdminOperationsHandler
	AdminFinance        *resellertransport.AdminFinanceHandler
}

func New(c *provider.Container) Handlers {
	directory := resellerAdminDirectoryAdapter{repo: c.ResellerRepo}
	return Handlers{
		User: resellertransport.NewUserHandler(
			c.ResellerManagementService,
			c.ResellerSiteConfigService,
			c.UploadService,
		),
		UserProductSetting: resellertransport.NewUserProductSettingHandler(c.ResellerProductSettingService),
		UserFinance:        resellertransport.NewUserFinanceHandler(c.ResellerAccountingService),
		UserOrder:          resellertransport.NewUserOrderHandler(c.ResellerOrderService),
		AdminManagement: resellertransport.NewAdminManagementHandler(
			c.ResellerManagementService, directory, c.AuthzAuditService,
		),
		AdminProfileDetail: resellertransport.NewAdminProfileDetailHandler(
			directory, c.ResellerProductSettingService, c.ResellerAccountingService, c.ResellerOrderService,
		),
		AdminSiteConfig: resellertransport.NewAdminSiteConfigHandler(
			c.ResellerSiteConfigService, directory, c.AuthzAuditService,
		),
		AdminProductSetting: resellertransport.NewAdminProductSettingHandler(
			c.ResellerProductSettingService, c.AuthzAuditService,
		),
		AdminOperations: resellertransport.NewAdminOperationsHandler(c.ResellerOperationsService),
		AdminFinance:    resellertransport.NewAdminFinanceHandler(c.ResellerAccountingService),
	}
}
