package router

import (
	affiliatebootstrap "github.com/dujiao-next/internal/bootstrap/affiliate"
	settingsbootstrap "github.com/dujiao-next/internal/bootstrap/settingshttp"
	"github.com/dujiao-next/internal/config"
	affiliatetransport "github.com/dujiao-next/internal/modules/affiliate/transport/http"
	channelclienthttp "github.com/dujiao-next/internal/modules/channelclient/transport/http"
	settingstransport "github.com/dujiao-next/internal/modules/settings/transport/http"
	broadcasthttp "github.com/dujiao-next/internal/modules/telegram/broadcast/transport/http"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/provider"
	adminauthtransport "github.com/dujiao-next/internal/transport/http/adminauth"
	adminauthztransport "github.com/dujiao-next/internal/transport/http/adminauthz"
	adminusertransport "github.com/dujiao-next/internal/transport/http/adminuser"
	adproxytransport "github.com/dujiao-next/internal/transport/http/adproxy"
	apicredentialtransport "github.com/dujiao-next/internal/transport/http/apicredential"
	auditlogtransport "github.com/dujiao-next/internal/transport/http/auditlog"
	cardsecrettransport "github.com/dujiao-next/internal/transport/http/cardsecret"
	catalogtransport "github.com/dujiao-next/internal/transport/http/catalog"
	compliancetransport "github.com/dujiao-next/internal/transport/http/compliance"
	contenttransport "github.com/dujiao-next/internal/transport/http/content"
	coupontransport "github.com/dujiao-next/internal/transport/http/coupon"
	dashboardtransport "github.com/dujiao-next/internal/transport/http/dashboard"
	fulfillmenttransport "github.com/dujiao-next/internal/transport/http/fulfillment"
	giftcardtransport "github.com/dujiao-next/internal/transport/http/giftcard"
	memberleveltransport "github.com/dujiao-next/internal/transport/http/memberlevel"
	notificationtransport "github.com/dujiao-next/internal/transport/http/notification"
	ordertransport "github.com/dujiao-next/internal/transport/http/order"
	paymenttransport "github.com/dujiao-next/internal/transport/http/payment"
	procurementtransport "github.com/dujiao-next/internal/transport/http/procurement"
	promotiontransport "github.com/dujiao-next/internal/transport/http/promotion"
	reconciliationtransport "github.com/dujiao-next/internal/transport/http/reconciliation"
	resellertransport "github.com/dujiao-next/internal/transport/http/reseller"
	siteconnectiontransport "github.com/dujiao-next/internal/transport/http/siteconnection"
	systemtransport "github.com/dujiao-next/internal/transport/http/system"
	uploadtransport "github.com/dujiao-next/internal/transport/http/upload"
	wallettransport "github.com/dujiao-next/internal/transport/http/wallet"
	adproxywiring "github.com/dujiao-next/internal/wiring/adproxy"
	compliancewiring "github.com/dujiao-next/internal/wiring/compliance"
	siteconnectionwiring "github.com/dujiao-next/internal/wiring/siteconnection"
	uploadwiring "github.com/dujiao-next/internal/wiring/upload"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func registerAdminRoutes(
	engine *gin.Engine,
	apiV1 *gin.RouterGroup,
	cfg *config.Config,
	c *provider.Container,
	adminLoginHandler *adminauthtransport.AdminLoginHandler,
	admin2FAHandler *adminauthtransport.Admin2FAHandler,
	adminUser2FAHandler *adminauthtransport.AdminUser2FAHandler,
	adminUserHandler *adminusertransport.AdminHandler,
	adminAuthzHandler *adminauthztransport.AdminHandler,
	adminFulfillmentHandler *fulfillmenttransport.AdminHandler,
	adminOrderHandler *ordertransport.AdminHandler,
	adminOrderRefundHandler *ordertransport.AdminRefundHandler,
	adminContentHandler *contenttransport.AdminHandler,
	adminDashboardHandler *dashboardtransport.AdminHandler,
	adminMemberLevelHandler *memberleveltransport.AdminHandler,
	adminApiCredentialHandler *apicredentialtransport.AdminHandler,
	adminAuditLogHandler *auditlogtransport.AdminHandler,
	adminCardSecretHandler *cardsecrettransport.AdminHandler,
	adminCatalogCategoryHandler *catalogtransport.AdminCategoryHandler,
	adminCatalogProductHandler *catalogtransport.AdminProductHandler,
	adminCatalogProductMappingHandler *catalogtransport.AdminProductMappingHandler,
	adminCouponHandler *coupontransport.AdminHandler,
	adminGiftCardHandler *giftcardtransport.AdminHandler,
	adminPromotionHandler *promotiontransport.AdminHandler,
	adminNotificationHandler *notificationtransport.AdminHandler,
	adminProcurementHandler *procurementtransport.AdminHandler,
	adminResellerManagementHandler *resellertransport.AdminManagementHandler,
	adminResellerProfileDetailHandler *resellertransport.AdminProfileDetailHandler,
	adminResellerSiteConfigHandler *resellertransport.AdminSiteConfigHandler,
	adminResellerProductSettingHandler *resellertransport.AdminProductSettingHandler,
	adminResellerOperationsHandler *resellertransport.AdminOperationsHandler,
	adminResellerFinanceHandler *resellertransport.AdminFinanceHandler,
	adminSettingsHandler *settingstransport.AdminHandler,
	adminWalletHandler *wallettransport.AdminHandler,
	adminPaymentHandler *paymenttransport.AdminHandler,
	adminPaymentChannelHandler *paymenttransport.AdminChannelHandler,
	redisClient *redis.Client,
	adminLoginRule RateLimitRule,
) {
	admin := apiV1.Group("/admin")

	// 登录接口（无需鉴权）
	adminauthtransport.RegisterAdminLoginAuthRoutes(admin, adminLoginHandler, RateLimitMiddleware(redisClient, adminLoginRule, KeyByIP))
	adminauthtransport.RegisterAdmin2FAAuthRoutes(admin, admin2FAHandler, RateLimitMiddleware(redisClient, adminLoginRule, KeyByIP))

	// 需要鉴权的接口
	authorized := admin.Use(JWTAuthMiddleware(cfg.JWT.SecretKey, c.AdminStore), AdminRBACMiddleware(c.AuthzService))
	// 支付/财务相关受保护子组：未确认合规声明时拦截
	// 注：admin.Use(...) 已 mutate admin 自身，新 Group 继承 JWT + RBAC 中间件
	paymentProtected := admin.Group("", PaymentComplianceRequired(c.ComplianceService))

	// 合规声明
	compliancetransport.RegisterAdminRoutes(authorized, compliancewiring.NewAdminHandler(c))

	// 仪表盘
	dashboardtransport.RegisterAdminRoutes(authorized, adminDashboardHandler)

	// 广告代理
	adproxytransport.RegisterAdminRoutes(authorized, adproxywiring.NewAdminHandler(c))

	// 商品 / 分类管理
	catalogtransport.RegisterAdminProductRoutes(authorized, adminCatalogProductHandler)
	contenttransport.RegisterAdminRoutes(authorized, adminContentHandler)
	catalogtransport.RegisterAdminCategoryRoutes(authorized, adminCatalogCategoryHandler)

	// 设置管理
	settingstransport.RegisterAdminRoutes(authorized, adminSettingsHandler)
	settingstransport.RegisterAdminSMTPRoutes(authorized, settingsbootstrap.NewSMTPHandler(c, cfg))
	settingstransport.RegisterAdminCaptchaRoutes(authorized, settingsbootstrap.NewCaptchaHandler(c, cfg))
	settingstransport.RegisterAdminTelegramAuthRoutes(authorized, settingsbootstrap.NewTelegramAuthHandler(c, cfg))
	notificationtransport.RegisterAdminRoutes(authorized, adminNotificationHandler)
	settingstransport.RegisterAdminOrderEmailTemplateRoutes(authorized, settingstransport.NewOrderEmailTemplateHandler(c.SettingService))
	settingstransport.RegisterAdminAffiliateRoutes(authorized, settingstransport.NewAffiliateHandler(c.SettingService))
	settingstransport.RegisterAdminTelegramBotRoutes(authorized, settingstransport.NewTelegramBotHandler(c.SettingService))
	adminauthtransport.RegisterAdminPasswordRoutes(authorized, adminLoginHandler)

	// 系统信息与版本检测
	systemtransport.RegisterAdminRoutes(authorized, systemtransport.NewAdminHandler(nil))

	adminauthtransport.RegisterAdmin2FARoutes(authorized, admin2FAHandler)

	// 推广返利
	adminAffiliateHandler := affiliatebootstrap.NewAdminHandler(c)
	affiliatetransport.RegisterAdminRoutes(authorized, adminAffiliateHandler)
	affiliatetransport.RegisterAdminFinanceRoutes(paymentProtected, adminAffiliateHandler)
	resellertransport.RegisterAdminOperationsOverviewRoutes(authorized, adminResellerOperationsHandler)
	resellertransport.RegisterAdminManagementRoutes(authorized, adminResellerManagementHandler)
	resellertransport.RegisterAdminProfileDetailRoutes(authorized, adminResellerProfileDetailHandler)
	resellertransport.RegisterAdminSiteConfigRoutes(authorized, adminResellerSiteConfigHandler)
	resellertransport.RegisterAdminProductSettingRoutes(authorized, adminResellerProductSettingHandler)
	resellertransport.RegisterAdminOperationsFinanceRoutes(paymentProtected, adminResellerOperationsHandler)
	resellertransport.RegisterAdminFinanceRoutes(paymentProtected, adminResellerFinanceHandler)

	// 权限管理
	adminauthztransport.RegisterAdminRoutes(authorized, adminAuthzHandler)
	auditlogtransport.RegisterAdminRoutes(authorized, adminAuditLogHandler)
	authorized.GET("/authz/permissions/catalog", func(ctx *gin.Context) {
		response.Success(ctx, buildAdminPermissionCatalog(engine))
	})

	// 文件上传
	uploadtransport.RegisterAdminRoutes(authorized, uploadwiring.NewAdminHandler(c))

	// 订单管理
	ordertransport.RegisterAdminRoutes(authorized, adminOrderHandler)
	ordertransport.RegisterAdminRefundWriteRoutes(authorized, adminOrderRefundHandler)
	ordertransport.RegisterAdminRefundRoutes(authorized, adminOrderRefundHandler)
	fulfillmenttransport.RegisterAdminRoutes(authorized, adminFulfillmentHandler)
	cardsecrettransport.RegisterAdminRoutes(authorized, adminCardSecretHandler)
	giftcardtransport.RegisterAdminRoutes(authorized, adminGiftCardHandler)

	// 优惠券与活动价
	coupontransport.RegisterAdminRoutes(authorized, adminCouponHandler)
	promotiontransport.RegisterAdminRoutes(authorized, adminPromotionHandler)

	// 会员等级
	memberleveltransport.RegisterAdminRoutes(authorized, adminMemberLevelHandler)

	// 支付渠道与支付记录
	paymenttransport.RegisterAdminChannelRoutes(paymentProtected, adminPaymentChannelHandler)
	paymenttransport.RegisterAdminRoutes(paymentProtected, adminPaymentHandler)

	// 用户管理
	adminusertransport.RegisterAdminRoutes(authorized, adminUserHandler)
	wallettransport.RegisterAdminRoutes(paymentProtected, adminWalletHandler)
	adminauthtransport.RegisterAdminUser2FARoutes(authorized, adminUser2FAHandler)

	// API 凭证审核管理
	apicredentialtransport.RegisterAdminRoutes(authorized, adminApiCredentialHandler)

	// 站点对接连接管理
	siteconnectiontransport.RegisterAdminRoutes(authorized, siteconnectionwiring.NewAdminHandler(c))

	// 商品映射管理
	catalogtransport.RegisterAdminProductMappingRoutes(authorized, adminCatalogProductMappingHandler)

	// 采购单管理
	procurementtransport.RegisterAdminRoutes(authorized, adminProcurementHandler)

	// 对账管理
	reconciliationtransport.RegisterAdminRoutes(paymentProtected, reconciliationtransport.NewAdminHandler(c.ReconciliationService))

	// 渠道客户端管理
	channelclienthttp.RegisterAdminRoutes(authorized, channelclienthttp.NewAdminHandler(c.ChannelClientService))

	// Telegram Bot 群发
	broadcasthttp.RegisterAdminRoutes(authorized, broadcasthttp.NewAdminHandler(c.TelegramBroadcastService))
}
