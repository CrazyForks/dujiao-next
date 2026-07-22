package resellerhttp

import "github.com/gin-gonic/gin"

// RegisterUserConsoleRoutes 注册用户中心分销商控制台（入驻/域名/站点配置）路由。
// 调用方必须传入已挂载登录鉴权与 RequireMainTenantForResellerConsole 的 RouterGroup。
func RegisterUserConsoleRoutes(console gin.IRoutes, handler *UserHandler) {
	if console == nil || handler == nil {
		panic("reseller user console routes: required dependency is nil")
	}
	console.GET("/profile", handler.GetManagementSnapshot)
	console.POST("/apply", handler.ApplyProfile)
	console.GET("/domains", handler.ListDomains)
	console.POST("/domains", handler.SubmitCustomDomain)
	console.GET("/site-config", handler.GetSiteConfig)
	console.PUT("/site-config", handler.UpdateSiteConfig)
	console.POST("/upload", handler.UploadImage)
}

// RegisterUserProductSettingRoutes 注册用户中心分销商品配置路由。
func RegisterUserProductSettingRoutes(console gin.IRoutes, handler *UserProductSettingHandler) {
	if console == nil || handler == nil {
		panic("reseller user product setting routes: required dependency is nil")
	}
	console.GET("/product-settings", handler.ListProductSettings)
	console.GET("/product-settings/:product_id", handler.GetProductSetting)
	console.POST("/product-settings/:product_id/preview", handler.PreviewProductSettings)
	console.PUT("/product-settings/:product_id", handler.UpdateProductSettings)
	console.DELETE("/product-settings/:product_id", handler.ResetProductSetting)
}

// RegisterUserFinanceRoutes 注册用户中心分销财务路由。
func RegisterUserFinanceRoutes(console gin.IRoutes, handler *UserFinanceHandler) {
	if console == nil || handler == nil {
		panic("reseller user finance routes: required dependency is nil")
	}
	console.GET("/dashboard", handler.GetDashboard)
	console.GET("/balance-accounts", handler.ListBalanceAccounts)
	console.GET("/ledger-entries", handler.ListLedgerEntries)
	console.GET("/withdraws", handler.ListWithdraws)
	console.POST("/withdraws", handler.ApplyWithdraw)
}

// RegisterUserOrderRoutes 注册用户中心分销销售订单只读路由。
func RegisterUserOrderRoutes(console gin.IRoutes, handler *UserOrderHandler) {
	if console == nil || handler == nil {
		panic("reseller user order routes: required dependency is nil")
	}
	console.GET("/orders", handler.ListOrders)
	console.GET("/orders/stats", handler.GetOrderStats)
	console.GET("/orders/:order_no", handler.GetOrderDetail)
}

// RegisterAdminManagementRoutes 注册后台分销商入驻/域名管理路由。
// 调用方必须传入已挂载认证与 RBAC 的 RouterGroup。
func RegisterAdminManagementRoutes(admin gin.IRoutes, handler *AdminManagementHandler) {
	if admin == nil || handler == nil {
		panic("reseller admin management routes: required dependency is nil")
	}
	admin.GET("/resellers/profiles", handler.ListProfiles)
	admin.PUT("/resellers/profiles/:id", handler.UpdateProfile)
	admin.PUT("/resellers/profiles/:id/system-domain", handler.AssignSystemDomain)
	admin.POST("/resellers/profiles/:id/approve", handler.ApproveProfile)
	admin.POST("/resellers/profiles/:id/reject", handler.RejectProfile)
	admin.POST("/resellers/profiles/:id/disable", handler.DisableProfile)
	admin.POST("/resellers/profiles/:id/restore", handler.RestoreProfile)
	admin.GET("/resellers/domains", handler.ListDomains)
	admin.POST("/resellers/domains/:id/approve", handler.ApproveDomain)
	admin.POST("/resellers/domains/:id/disable", handler.DisableDomain)
	admin.POST("/resellers/domains/:id/set-primary", handler.SetPrimaryDomain)
}

// RegisterAdminProfileDetailRoutes 注册后台分销商运营详情路由。
func RegisterAdminProfileDetailRoutes(admin gin.IRoutes, handler *AdminProfileDetailHandler) {
	if admin == nil || handler == nil {
		panic("reseller admin profile detail routes: required dependency is nil")
	}
	admin.GET("/resellers/profiles/:id", handler.GetProfileDetail)
}

// RegisterAdminSiteConfigRoutes 注册后台分销站点配置路由。
func RegisterAdminSiteConfigRoutes(admin gin.IRoutes, handler *AdminSiteConfigHandler) {
	if admin == nil || handler == nil {
		panic("reseller admin site config routes: required dependency is nil")
	}
	admin.GET("/resellers/site-configs", handler.ListSiteConfigs)
	admin.GET("/resellers/site-configs/:reseller_id", handler.GetSiteConfig)
	admin.PUT("/resellers/site-configs/:reseller_id", handler.UpdateSiteConfig)
	admin.POST("/resellers/site-configs/:reseller_id/reset", handler.ResetSiteConfig)
}

// RegisterAdminProductSettingRoutes 注册后台分销商品配置路由。
func RegisterAdminProductSettingRoutes(admin gin.IRoutes, handler *AdminProductSettingHandler) {
	if admin == nil || handler == nil {
		panic("reseller admin product setting routes: required dependency is nil")
	}
	admin.GET("/resellers/product-settings", handler.ListProductSettings)
	admin.GET("/resellers/product-settings/:reseller_id/:product_id", handler.GetProductSetting)
	admin.POST("/resellers/product-settings/:reseller_id/:product_id/preview", handler.PreviewProductSettings)
	admin.PUT("/resellers/product-settings/:reseller_id/:product_id", handler.UpdateProductSettings)
	admin.DELETE("/resellers/product-settings/:reseller_id/:product_id", handler.ResetProductSetting)
}

// RegisterAdminOperationsOverviewRoutes 注册后台分销运营 overview 路由（普通授权组）。
func RegisterAdminOperationsOverviewRoutes(admin gin.IRoutes, handler *AdminOperationsHandler) {
	if admin == nil || handler == nil {
		panic("reseller admin operations overview routes: required dependency is nil")
	}
	admin.GET("/resellers/operations/overview", handler.GetOverview)
}

// RegisterAdminOperationsFinanceRoutes 注册后台分销运营 finance 路由（须挂在支付合规保护组）。
func RegisterAdminOperationsFinanceRoutes(admin gin.IRoutes, handler *AdminOperationsHandler) {
	if admin == nil || handler == nil {
		panic("reseller admin operations finance routes: required dependency is nil")
	}
	admin.GET("/resellers/operations/finance", handler.GetFinance)
}

// RegisterAdminFinanceRoutes 注册后台分销财务路由（须挂在支付合规保护组）。
func RegisterAdminFinanceRoutes(admin gin.IRoutes, handler *AdminFinanceHandler) {
	if admin == nil || handler == nil {
		panic("reseller admin finance routes: required dependency is nil")
	}
	admin.GET("/resellers/ledger-entries", handler.ListLedgerEntries)
	admin.GET("/resellers/balance-accounts", handler.ListBalanceAccounts)
	admin.GET("/resellers/withdraws", handler.ListWithdraws)
	admin.POST("/resellers/withdraws/:id/reject", handler.RejectWithdraw)
	admin.POST("/resellers/withdraws/:id/pay", handler.PayWithdraw)
}
