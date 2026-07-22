package router

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSetupRouterDelegatesRouteDomains(t *testing.T) {
	routerDirectory := currentRouterDirectory(t)
	source := readRouterSource(t, filepath.Join(routerDirectory, "router.go"))

	for _, registration := range []string{
		"registerStorefrontRoutes(",
		"registerUpstreamRoutes(",
		"registerChannelRoutes(",
		"registerPaymentCallbackRoutes(",
		"registerAdminRoutes(",
	} {
		if !strings.Contains(source, registration) {
			t.Errorf("SetupRouter must delegate through %s", registration)
		}
	}

	for _, inlineGroup := range []string{
		`apiV1.Group("/upstream")`,
		`apiV1.Group("/channel")`,
		`apiV1.Group("/admin")`,
	} {
		if strings.Contains(source, inlineGroup) {
			t.Errorf("router.go must not retain inline domain group %s", inlineGroup)
		}
	}
}

func TestRouteDomainFilesPreserveTrustBoundaries(t *testing.T) {
	routerDirectory := currentRouterDirectory(t)
	tests := []struct {
		file     string
		required []string
	}{
		{
			file: "routes_storefront.go",
			required: []string{
				`storefront.Use(ResellerTenantMiddleware(`,
				`catalogtransport.RegisterPublicRoutes(public, publicCatalogHandler)`,
				`contenttransport.RegisterPublicRoutes(public, publicContentHandler)`,
				`captchatransport.RegisterPublicRoutes(public,`,
				`affiliatetransport.RegisterPublicRoutes(public, affiliateHandler)`,
				`affiliatetransport.RegisterUserRoutes(user, affiliateHandler)`,
				`user.Use(UserJWTAuthMiddleware(`,
				`resellerConsole.Use(RequireMainTenantForResellerConsole())`,
				`resellertransport.RegisterUserConsoleRoutes(resellerConsole, userResellerHandler)`,
				`resellertransport.RegisterUserProductSettingRoutes(resellerConsole, userResellerProductSettingHandler)`,
				`resellertransport.RegisterUserFinanceRoutes(resellerConsole, userResellerFinanceHandler)`,
				`resellertransport.RegisterUserOrderRoutes(resellerConsole, userResellerOrderHandler)`,
				`apicredentialtransport.RegisterUserRoutes(user, userApiCredentialHandler)`,
				`auditlogtransport.RegisterUserRoutes(user, userAuditLogHandler)`,
				`giftcardtransport.RegisterUserRoutes(user, userGiftCardHandler)`,
				`wallettransport.RegisterUserRoutes(user, userWalletHandler)`,
				`userauthtransport.RegisterUserProfileRoutes(user, userProfileHandler)`,
				`userauthtransport.RegisterUserEmailRoutes(user, userEmailHandler)`,
				`userauthtransport.RegisterUserVerifyAuthRoutes(auth, userVerifyHandler)`,
				`userauthtransport.RegisterUserRegisterAuthRoutes(auth, userLoginHandler)`,
				`userauthtransport.RegisterUserLoginAuthRoutes(auth, userLoginHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIPAndJSONField("email")))`,
				`userauthtransport.RegisterUser2FAAuthRoutes(auth, user2FAHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIP))`,
				`userauthtransport.RegisterUser2FARoutes(user, user2FAHandler)`,
				`userauthtransport.RegisterUserTelegramAuthRoutes(auth, userTelegramHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIP))`,
				`userauthtransport.RegisterUserTelegramRoutes(user, userTelegramHandler)`,
				`userauthtransport.RegisterUserTelegramOIDCAuthRoutes(auth, userTelegramOIDCHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIP))`,
				`userauthtransport.RegisterUserTelegramOIDCRoutes(user, userTelegramOIDCHandler)`,
				`userauthtransport.RegisterUserPasswordAuthRoutes(auth, userPasswordHandler)`,
				`userauthtransport.RegisterUserPasswordRoutes(user, userPasswordHandler)`,
				`memberleveltransport.RegisterPublicRoutes(public, publicMemberLevelHandler)`,
				`publicconfigtransport.RegisterPublicRoutes(public, publicConfigHandler)`,
				`carttransport.RegisterUserRoutes(user, userCartHandler)`,
				`ordertransport.RegisterUserReadRoutes(user, userOrderHandler)`,
				`ordertransport.RegisterUserCancelRoute(user, userOrderHandler)`,
				`ordertransport.RegisterUserPreviewRoute(user, orderPreviewHandler)`,
				`ordertransport.RegisterUserCreateRoute(user, orderCreateHandler)`,
				`ordertransport.RegisterUserCreateAndPayRoute(user, orderCreateHandler)`,
				`ordertransport.RegisterUserPaymentChannelsRoute(user, userOrderHandler)`,
				`ordertransport.RegisterGuestReadRoutes(guest, guestOrderHandler)`,
				`ordertransport.RegisterGuestPreviewRoute(guest, orderPreviewHandler)`,
				`ordertransport.RegisterGuestCreateRoute(guest, orderCreateHandler)`,
				`ordertransport.RegisterGuestCreateAndPayRoute(guest, orderCreateHandler)`,
				`paymenttransport.RegisterGuestWriteRoutes(guest, paymentWriteHandler)`,
				`paymenttransport.RegisterGuestLatestRoute(guest, paymentLatestHandler)`,
				`paymenttransport.RegisterUserWriteRoutes(user, paymentWriteHandler)`,
				`paymenttransport.RegisterUserLatestRoute(user, paymentLatestHandler)`,
				`paymenttransport.RegisterWebhookRoutes(apiV1, webhookHandler)`,
			},
		},
		{
			file: "routes_upstream.go",
			required: []string{
				`upstreamAPI.Use(RateLimitMiddleware(`,
				`upstreamAPI.Use(UpstreamAPIAuthMiddleware(`,
				`apiV1.POST("/upstream/callback",`,
			},
		},
		{
			file: "routes_channel.go",
			required: []string{
				`channelAPI.Use(ChannelAPIAuthMiddleware(`,
				`telegramtransport.RegisterChannelBotRoutes(channelAPI, channelTelegramBotHandler)`,
				`affiliatetransport.RegisterChannelRoutes(channelAPI, channelAffiliateHandler)`,
				`memberleveltransport.RegisterChannelRoutes(channelAPI, channelMemberLevelHandler)`,
				`wallettransport.RegisterChannelRoutes(channelAPI, channelWalletHandler)`,
				`giftcardtransport.RegisterChannelRoutes(channelAPI, channelGiftCardHandler)`,
			},
		},
		{
			file: "routes_admin.go",
			required: []string{
				`adminauthtransport.RegisterAdminLoginAuthRoutes(admin, adminLoginHandler, RateLimitMiddleware(redisClient, adminLoginRule, KeyByIP))`,
				`adminauthtransport.RegisterAdmin2FAAuthRoutes(admin, admin2FAHandler, RateLimitMiddleware(redisClient, adminLoginRule, KeyByIP))`,
				`adminauthtransport.RegisterAdminPasswordRoutes(authorized, adminLoginHandler)`,
				`adminauthtransport.RegisterAdmin2FARoutes(authorized, admin2FAHandler)`,
				`adminauthtransport.RegisterAdminUser2FARoutes(authorized, adminUser2FAHandler)`,
				`adminusertransport.RegisterAdminRoutes(authorized, adminUserHandler)`,
				`adminauthztransport.RegisterAdminRoutes(authorized, adminAuthzHandler)`,
				`ordertransport.RegisterAdminRoutes(authorized, adminOrderHandler)`,
				`ordertransport.RegisterAdminRefundWriteRoutes(authorized, adminOrderRefundHandler)`,
				`ordertransport.RegisterAdminRefundRoutes(authorized, adminOrderRefundHandler)`,
				`fulfillmenttransport.RegisterAdminRoutes(authorized, adminFulfillmentHandler)`,
				`authorized := admin.Use(JWTAuthMiddleware(`,
				`paymentProtected := admin.Group("", PaymentComplianceRequired(`,
				`compliancetransport.RegisterAdminRoutes(authorized,`,
				`systemtransport.RegisterAdminRoutes(authorized,`,
				`adproxytransport.RegisterAdminRoutes(authorized,`,
				`contenttransport.RegisterAdminRoutes(authorized, adminContentHandler)`,
				`dashboardtransport.RegisterAdminRoutes(authorized, adminDashboardHandler)`,
				`memberleveltransport.RegisterAdminRoutes(authorized, adminMemberLevelHandler)`,
				`apicredentialtransport.RegisterAdminRoutes(authorized, adminApiCredentialHandler)`,
				`auditlogtransport.RegisterAdminRoutes(authorized, adminAuditLogHandler)`,
				`cardsecrettransport.RegisterAdminRoutes(authorized, adminCardSecretHandler)`,
				`giftcardtransport.RegisterAdminRoutes(authorized, adminGiftCardHandler)`,
				`settingstransport.RegisterAdminRoutes(authorized, adminSettingsHandler)`,
				`settingstransport.RegisterAdminSMTPRoutes(authorized,`,
				`settingstransport.RegisterAdminCaptchaRoutes(authorized,`,
				`settingstransport.RegisterAdminTelegramAuthRoutes(authorized,`,
				`settingstransport.RegisterAdminOrderEmailTemplateRoutes(authorized,`,
				`settingstransport.RegisterAdminAffiliateRoutes(authorized,`,
				`settingstransport.RegisterAdminTelegramBotRoutes(authorized,`,
				`uploadtransport.RegisterAdminRoutes(authorized,`,
				`broadcasthttp.RegisterAdminRoutes(authorized,`,
				`channelclienttransport.RegisterAdminRoutes(authorized,`,
				`siteconnectiontransport.RegisterAdminRoutes(authorized,`,
				`affiliatetransport.RegisterAdminRoutes(authorized, adminAffiliateHandler)`,
				`affiliatetransport.RegisterAdminFinanceRoutes(paymentProtected, adminAffiliateHandler)`,
				`catalogtransport.RegisterAdminProductRoutes(authorized, adminCatalogProductHandler)`,
				`catalogtransport.RegisterAdminCategoryRoutes(authorized, adminCatalogCategoryHandler)`,
				`catalogtransport.RegisterAdminProductMappingRoutes(authorized, adminCatalogProductMappingHandler)`,
				`coupontransport.RegisterAdminRoutes(authorized, adminCouponHandler)`,
				`promotiontransport.RegisterAdminRoutes(authorized, adminPromotionHandler)`,
				`notificationtransport.RegisterAdminRoutes(authorized, adminNotificationHandler)`,
				`procurementtransport.RegisterAdminRoutes(authorized, adminProcurementHandler)`,
				`resellertransport.RegisterAdminOperationsOverviewRoutes(authorized, adminResellerOperationsHandler)`,
				`resellertransport.RegisterAdminManagementRoutes(authorized, adminResellerManagementHandler)`,
				`resellertransport.RegisterAdminProfileDetailRoutes(authorized, adminResellerProfileDetailHandler)`,
				`resellertransport.RegisterAdminSiteConfigRoutes(authorized, adminResellerSiteConfigHandler)`,
				`resellertransport.RegisterAdminProductSettingRoutes(authorized, adminResellerProductSettingHandler)`,
				`resellertransport.RegisterAdminOperationsFinanceRoutes(paymentProtected, adminResellerOperationsHandler)`,
				`resellertransport.RegisterAdminFinanceRoutes(paymentProtected, adminResellerFinanceHandler)`,
				`paymenttransport.RegisterAdminChannelRoutes(paymentProtected, adminPaymentChannelHandler)`,
				`paymenttransport.RegisterAdminRoutes(paymentProtected, adminPaymentHandler)`,
				`wallettransport.RegisterAdminRoutes(paymentProtected, adminWalletHandler)`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			source := readRouterSource(t, filepath.Join(routerDirectory, test.file))
			for _, required := range test.required {
				if !strings.Contains(source, required) {
					t.Errorf("%s must preserve trust-boundary statement %q", test.file, required)
				}
			}
		})
	}
}

func currentRouterDirectory(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve router test filename")
	}
	return filepath.Dir(thisFile)
}

func readRouterSource(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read router source %s: %v", path, err)
	}
	return string(raw)
}
