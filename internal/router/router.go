package router

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/dujiao-next/internal/authz"
	affiliatebootstrap "github.com/dujiao-next/internal/bootstrap/affiliate"
	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	captchahttp "github.com/dujiao-next/internal/modules/captcha/transport/http"
	settingstransport "github.com/dujiao-next/internal/modules/settings/transport/http"
	"github.com/dujiao-next/internal/provider"
	apicredentialtransport "github.com/dujiao-next/internal/transport/http/apicredential"
	auditlogtransport "github.com/dujiao-next/internal/transport/http/auditlog"
	cardsecrettransport "github.com/dujiao-next/internal/transport/http/cardsecret"
	catalogtransport "github.com/dujiao-next/internal/transport/http/catalog"
	contenttransport "github.com/dujiao-next/internal/transport/http/content"
	coupontransport "github.com/dujiao-next/internal/transport/http/coupon"
	dashboardtransport "github.com/dujiao-next/internal/transport/http/dashboard"
	giftcardtransport "github.com/dujiao-next/internal/transport/http/giftcard"
	memberleveltransport "github.com/dujiao-next/internal/transport/http/memberlevel"
	notificationtransport "github.com/dujiao-next/internal/transport/http/notification"
	procurementtransport "github.com/dujiao-next/internal/transport/http/procurement"
	promotiontransport "github.com/dujiao-next/internal/transport/http/promotion"
	sitemaptransport "github.com/dujiao-next/internal/transport/http/sitemap"
	"github.com/dujiao-next/internal/web"
	adminauthwiring "github.com/dujiao-next/internal/wiring/adminauth"
	adminauthzwiring "github.com/dujiao-next/internal/wiring/adminauthz"
	adminuserwiring "github.com/dujiao-next/internal/wiring/adminuser"
	cartwiring "github.com/dujiao-next/internal/wiring/cart"
	catalogwiring "github.com/dujiao-next/internal/wiring/catalog"
	channelwiring "github.com/dujiao-next/internal/wiring/channel"
	channeluserwiring "github.com/dujiao-next/internal/wiring/channeluser"
	fulfillmentwiring "github.com/dujiao-next/internal/wiring/fulfillment"
	orderwiring "github.com/dujiao-next/internal/wiring/order"
	paymentwiring "github.com/dujiao-next/internal/wiring/payment"
	publicconfigwiring "github.com/dujiao-next/internal/wiring/publicconfig"
	resellerwiring "github.com/dujiao-next/internal/wiring/reseller"
	sitemapwiring "github.com/dujiao-next/internal/wiring/sitemap"
	telegramwiring "github.com/dujiao-next/internal/wiring/telegram"
	upstreamwiring "github.com/dujiao-next/internal/wiring/upstream"
	userauthwiring "github.com/dujiao-next/internal/wiring/userauth"
	walletwiring "github.com/dujiao-next/internal/wiring/wallet"

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化路由。
func SetupRouter(cfg *config.Config, c *provider.Container) *gin.Engine {
	log := logger.L
	if log == nil {
		log = logger.Init(cfg.Server.Mode, cfg.Log.ToLoggerOptions())
	}
	r := gin.New()
	captchaVerifier := captchahttp.NewVerifier(c.CaptchaService)

	// 初始化 Handler（按前台/后台分组）
	adminAuthHandlers := adminauthwiring.New(c)
	adminLoginHandler := adminAuthHandlers.Login
	admin2FAHandler := adminAuthHandlers.TwoFA
	adminUser2FAHandler := adminAuthHandlers.UserTwoFA
	adminUserHandler := adminuserwiring.NewHandler(c)
	adminAuthzHandler := adminauthzwiring.NewHandler(c)
	adminFulfillmentHandler := fulfillmentwiring.NewAdminHandler(c)
	orderHandlers := orderwiring.New(c)
	adminOrderHandler := orderHandlers.Admin
	adminOrderRefundHandler := orderHandlers.AdminRefund
	userOrderHandler := orderHandlers.User
	guestOrderHandler := orderHandlers.Guest
	orderPreviewHandler := orderHandlers.Preview
	orderCreateHandler := orderHandlers.Create
	paymentHandlers := paymentwiring.New(c)
	paymentLatestHandler := paymentHandlers.Latest
	paymentWriteHandler := paymentHandlers.Write
	adminPaymentHandler := paymentHandlers.Admin
	adminPaymentChannelHandler := paymentHandlers.AdminChannel
	paymentWebhookHandler := paymentHandlers.Webhook
	paymentCallbackHandler := paymentHandlers.Callback
	publicConfigHandler := publicconfigwiring.NewHandler(c)
	userCartHandler := cartwiring.NewUserHandler(c)
	channelHandler := channelwiring.NewHandler(c)
	upstreamHandler := upstreamwiring.NewHandler(c)
	publicContentHandler := contenttransport.NewPublicHandler(
		c.ContentPostService,
		c.ContentPostCategoryService,
		c.ContentBannerService,
	)
	publicCatalogHandler := catalogwiring.NewPublicHandler(c)
	adminContentHandler := contenttransport.NewAdminHandler(
		c.ContentPostService,
		c.ContentPostCategoryService,
		c.ContentBannerService,
		c.ContentMediaService,
	)
	adminDashboardHandler := dashboardtransport.NewAdminHandler(c.DashboardService)
	adminMemberLevelHandler := memberleveltransport.NewAdminHandler(c.MemberLevelService)
	publicMemberLevelHandler := memberleveltransport.NewPublicHandler(c.MemberLevelService)
	userAuthHandlers := userauthwiring.New(c)
	userProfileHandler := userAuthHandlers.Profile
	userEmailHandler := userAuthHandlers.Email
	userPasswordHandler := userAuthHandlers.Password
	userVerifyHandler := userAuthHandlers.Verify
	userLoginHandler := userAuthHandlers.Login
	user2FAHandler := userAuthHandlers.TwoFA
	userTelegramOIDCHandler := userAuthHandlers.TelegramOIDC
	userTelegramHandler := userAuthHandlers.Telegram
	walletHandlers := walletwiring.New(c)
	userWalletHandler := walletHandlers.User
	adminWalletHandler := walletHandlers.Admin
	channelWalletHandler := walletHandlers.Channel
	channelMemberLevelHandler := memberleveltransport.NewChannelHandler(c.MemberLevelService)
	adminApiCredentialHandler := apicredentialtransport.NewAdminHandler(c.ApiCredentialService)
	userApiCredentialHandler := apicredentialtransport.NewUserHandler(c.ApiCredentialService)
	adminAuditLogHandler := auditlogtransport.NewAdminHandler(c.AuthzAuditService, c.UserLoginLogService)
	adminCardSecretHandler := cardsecrettransport.NewAdminHandler(c.CardSecretService)
	adminCatalogCategoryHandler := catalogtransport.NewAdminCategoryHandler(c.CategoryService)
	adminCatalogProductHandler := catalogtransport.NewAdminProductHandler(
		c.ProductService,
		c.ProductService,
		c.SettingService,
		c.ProductMappingRepo,
		c.SKUMappingRepo,
	)
	adminCatalogProductMappingHandler := catalogtransport.NewAdminProductMappingHandler(c.ProductMappingService.Service)
	userAuditLogHandler := auditlogtransport.NewUserHandler(c.UserLoginLogService)
	adminCouponHandler := coupontransport.NewAdminHandler(c.CouponAdminService)
	adminGiftCardHandler := giftcardtransport.NewAdminHandler(c.GiftCardService)
	userGiftCardHandler := giftcardtransport.NewUserHandler(c.GiftCardService, captchaVerifier)
	channelGiftCardHandler := giftcardtransport.NewChannelHandler(
		c.GiftCardService,
		channeluserwiring.NewSimpleProvisioner(c.UserAuthService),
	)
	channelAffiliateHandler := affiliatebootstrap.NewChannelHandler(c)
	channelTelegramBotHandler := telegramwiring.NewChannelBotHandler(c)
	adminSettingsHandler := settingstransport.NewAdminHandler(c.SettingService)
	adminPromotionHandler := promotiontransport.NewAdminHandler(c.PromotionAdminService)
	adminNotificationHandler := notificationtransport.NewAdminHandler(c.SettingService, c.NotificationLogService, c.NotificationService)
	adminProcurementHandler := procurementtransport.NewAdminHandler(c.ProcurementOrderService)
	resellerHandlers := resellerwiring.New(c)
	userResellerHandler := resellerHandlers.User
	userResellerProductSettingHandler := resellerHandlers.UserProductSetting
	userResellerFinanceHandler := resellerHandlers.UserFinance
	userResellerOrderHandler := resellerHandlers.UserOrder
	adminResellerManagementHandler := resellerHandlers.AdminManagement
	adminResellerProfileDetailHandler := resellerHandlers.AdminProfileDetail
	adminResellerSiteConfigHandler := resellerHandlers.AdminSiteConfig
	adminResellerProductSettingHandler := resellerHandlers.AdminProductSetting
	adminResellerOperationsHandler := resellerHandlers.AdminOperations
	adminResellerFinanceHandler := resellerHandlers.AdminFinance

	redisPrefix := strings.TrimSpace(cfg.Redis.Prefix)
	if redisPrefix == "" {
		redisPrefix = constants.RedisPrefixDefault
	}
	redisClient := cache.Client()
	loginRule := RateLimitRule{
		Prefix:        fmt.Sprintf("%s:rate:login", redisPrefix),
		WindowSeconds: cfg.Security.LoginRateLimit.WindowSeconds,
		MaxRequests:   cfg.Security.LoginRateLimit.MaxAttempts,
		BlockSeconds:  cfg.Security.LoginRateLimit.BlockSeconds,
		MessageKey:    "error.login_too_many",
	}
	adminLoginRule := RateLimitRule{
		Prefix:        fmt.Sprintf("%s:rate:admin_login", redisPrefix),
		WindowSeconds: cfg.Security.LoginRateLimit.WindowSeconds,
		MaxRequests:   cfg.Security.LoginRateLimit.MaxAttempts,
		BlockSeconds:  cfg.Security.LoginRateLimit.BlockSeconds,
		MessageKey:    "error.login_too_many",
	}
	upstreamAPIRule := RateLimitRule{
		Prefix:        fmt.Sprintf("%s:rate:upstream_api", redisPrefix),
		WindowSeconds: 60,
		MaxRequests:   60,
		BlockSeconds:  30,
		MessageKey:    "error.rate_limited",
	}

	// RequestIDMiddleware 必须前置于 RecoveryMiddleware：panic 日志与响应都依赖 request_id。
	r.Use(RequestIDMiddleware())
	r.Use(RecoveryMiddleware())
	r.Use(LoggerMiddleware(log))
	r.Use(CORSMiddleware(cfg.CORS))
	r.Use(CallbackRouteMiddleware(c.SettingService, paymentCallbackHandler, paymentWebhookHandler, upstreamHandler))

	// 静态文件服务（上传的图片）必须放在前面。
	r.Static("/uploads", "./uploads")

	// SEO 资源（动态生成）。
	sitemaptransport.RegisterRoutes(r, sitemapwiring.NewHandler(c))

	apiV1 := r.Group("/api/v1")
	registerStorefrontRoutes(apiV1, cfg, c, publicContentHandler, publicCatalogHandler, userResellerHandler, userResellerProductSettingHandler, userResellerFinanceHandler, userResellerOrderHandler, userApiCredentialHandler, userAuditLogHandler, userGiftCardHandler, publicMemberLevelHandler, userProfileHandler, userEmailHandler, userPasswordHandler, userVerifyHandler, userTelegramOIDCHandler, userTelegramHandler, userLoginHandler, user2FAHandler, publicConfigHandler, userCartHandler, userOrderHandler, guestOrderHandler, orderPreviewHandler, orderCreateHandler, paymentLatestHandler, paymentWriteHandler, userWalletHandler, redisClient, loginRule)
	registerUpstreamRoutes(apiV1, c, upstreamHandler, redisClient, upstreamAPIRule)
	registerChannelRoutes(apiV1, c, channelHandler, channelMemberLevelHandler, channelGiftCardHandler, channelAffiliateHandler, channelTelegramBotHandler, channelWalletHandler)
	registerPaymentCallbackRoutes(apiV1, paymentCallbackHandler, paymentWebhookHandler)
	registerAdminRoutes(r, apiV1, cfg, c, adminLoginHandler, admin2FAHandler, adminUser2FAHandler, adminUserHandler, adminAuthzHandler, adminFulfillmentHandler, adminOrderHandler, adminOrderRefundHandler, adminContentHandler, adminDashboardHandler, adminMemberLevelHandler, adminApiCredentialHandler, adminAuditLogHandler, adminCardSecretHandler, adminCatalogCategoryHandler, adminCatalogProductHandler, adminCatalogProductMappingHandler, adminCouponHandler, adminGiftCardHandler, adminPromotionHandler, adminNotificationHandler, adminProcurementHandler, adminResellerManagementHandler, adminResellerProfileDetailHandler, adminResellerSiteConfigHandler, adminResellerProductSettingHandler, adminResellerOperationsHandler, adminResellerFinanceHandler, adminSettingsHandler, adminWalletHandler, adminPaymentHandler, adminPaymentChannelHandler, redisClient, adminLoginRule)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 嵌入式前端资源（仅在 -tags fullstack 构建时生效）
	if web.Enabled() {
		if err := web.ValidateAdminPath(cfg.Web.AdminPath); err != nil {
			log.Sugar().Fatalf("web.admin_path 配置错误: %v", err)
		}
		if err := web.RegisterAdmin(r, cfg.Web.AdminPath, web.AdminFS()); err != nil {
			log.Sugar().Fatalf("注册 admin SPA 失败: %v", err)
		}
		if err := web.RegisterUser(r, web.UserFS()); err != nil {
			log.Sugar().Fatalf("注册 user SPA 失败: %v", err)
		}
	}

	return r
}

type adminPermissionCatalogItem struct {
	Module     string `json:"module"`
	Method     string `json:"method"`
	Object     string `json:"object"`
	Permission string `json:"permission"`
}

func buildAdminPermissionCatalog(engine *gin.Engine) []adminPermissionCatalogItem {
	if engine == nil {
		return []adminPermissionCatalogItem{}
	}

	routes := engine.Routes()
	seen := make(map[string]struct{}, len(routes))
	items := make([]adminPermissionCatalogItem, 0, len(routes))

	for _, item := range routes {
		method := strings.ToUpper(strings.TrimSpace(item.Method))
		if method == "" || method == http.MethodOptions || method == http.MethodHead {
			continue
		}
		if !strings.HasPrefix(item.Path, "/api/v1/admin/") {
			continue
		}
		if item.Path == "/api/v1/admin/login" {
			continue
		}
		if item.Path == "/api/v1/admin/login/verify-2fa" {
			continue
		}
		object := authz.NormalizeObject(item.Path)
		permission := method + ":" + object
		if _, exists := seen[permission]; exists {
			continue
		}
		seen[permission] = struct{}{}
		items = append(items, adminPermissionCatalogItem{
			Module:     deriveAdminPermissionModule(object),
			Method:     method,
			Object:     object,
			Permission: permission,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Module == items[j].Module {
			if items[i].Object == items[j].Object {
				return items[i].Method < items[j].Method
			}
			return items[i].Object < items[j].Object
		}
		return items[i].Module < items[j].Module
	})

	return items
}

func deriveAdminPermissionModule(object string) string {
	normalized := strings.TrimPrefix(strings.TrimSpace(object), "/")
	if normalized == "" {
		return "system"
	}
	segments := strings.Split(normalized, "/")
	if len(segments) <= 1 {
		return segments[0]
	}
	if segments[0] != "admin" {
		return segments[0]
	}
	if segments[1] == "authz" {
		return "authz"
	}
	return segments[1]
}
