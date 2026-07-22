package router

import (
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/provider"
	affiliatetransport "github.com/dujiao-next/internal/transport/http/affiliate"
	apicredentialtransport "github.com/dujiao-next/internal/transport/http/apicredential"
	auditlogtransport "github.com/dujiao-next/internal/transport/http/auditlog"
	captchatransport "github.com/dujiao-next/internal/transport/http/captcha"
	carttransport "github.com/dujiao-next/internal/transport/http/cart"
	catalogtransport "github.com/dujiao-next/internal/transport/http/catalog"
	contenttransport "github.com/dujiao-next/internal/transport/http/content"
	giftcardtransport "github.com/dujiao-next/internal/transport/http/giftcard"
	memberleveltransport "github.com/dujiao-next/internal/transport/http/memberlevel"
	ordertransport "github.com/dujiao-next/internal/transport/http/order"
	paymenttransport "github.com/dujiao-next/internal/transport/http/payment"
	paymentcallbacktransport "github.com/dujiao-next/internal/transport/http/payment/callback"
	publicconfigtransport "github.com/dujiao-next/internal/transport/http/publicconfig"
	resellertransport "github.com/dujiao-next/internal/transport/http/reseller"
	userauthtransport "github.com/dujiao-next/internal/transport/http/userauth"
	wallettransport "github.com/dujiao-next/internal/transport/http/wallet"
	affiliatewiring "github.com/dujiao-next/internal/wiring/affiliate"
	captchawiring "github.com/dujiao-next/internal/wiring/captcha"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func registerStorefrontRoutes(
	apiV1 *gin.RouterGroup,
	cfg *config.Config,
	c *provider.Container,
	publicContentHandler *contenttransport.PublicHandler,
	publicCatalogHandler *catalogtransport.PublicHandler,
	userResellerHandler *resellertransport.UserHandler,
	userResellerProductSettingHandler *resellertransport.UserProductSettingHandler,
	userResellerFinanceHandler *resellertransport.UserFinanceHandler,
	userResellerOrderHandler *resellertransport.UserOrderHandler,
	userApiCredentialHandler *apicredentialtransport.UserHandler,
	userAuditLogHandler *auditlogtransport.UserHandler,
	userGiftCardHandler *giftcardtransport.UserHandler,
	publicMemberLevelHandler *memberleveltransport.PublicHandler,
	userProfileHandler *userauthtransport.UserProfileHandler,
	userEmailHandler *userauthtransport.UserEmailHandler,
	userPasswordHandler *userauthtransport.UserPasswordHandler,
	userVerifyHandler *userauthtransport.UserVerifyHandler,
	userTelegramOIDCHandler *userauthtransport.UserTelegramOIDCHandler,
	userTelegramHandler *userauthtransport.UserTelegramHandler,
	userLoginHandler *userauthtransport.UserLoginHandler,
	user2FAHandler *userauthtransport.User2FAHandler,
	publicConfigHandler *publicconfigtransport.Handler,
	userCartHandler *carttransport.UserHandler,
	userOrderHandler *ordertransport.UserHandler,
	guestOrderHandler *ordertransport.GuestHandler,
	orderPreviewHandler *ordertransport.PreviewHandler,
	orderCreateHandler *ordertransport.CreateHandler,
	paymentLatestHandler *paymenttransport.LatestHandler,
	paymentWriteHandler *paymenttransport.WriteHandler,
	userWalletHandler *wallettransport.UserHandler,
	redisClient *redis.Client,
	loginRule RateLimitRule,
) {
	storefront := apiV1.Group("")
	storefront.Use(ResellerTenantMiddleware(c.ResellerDomainResolver))
	affiliateHandler := affiliatewiring.NewStorefrontHandler(c)

	// 公开接口
	public := storefront.Group("/public")
	{
		publicconfigtransport.RegisterPublicRoutes(public, publicConfigHandler)
		catalogtransport.RegisterPublicRoutes(public, publicCatalogHandler)
		contenttransport.RegisterPublicRoutes(public, publicContentHandler)
		captchatransport.RegisterPublicRoutes(public, captchawiring.NewPublicHandler(c))
		affiliatetransport.RegisterPublicRoutes(public, affiliateHandler)
		memberleveltransport.RegisterPublicRoutes(public, publicMemberLevelHandler)
	}

	// 游客接口
	guest := storefront.Group("/guest")
	{
		ordertransport.RegisterGuestCreateRoute(guest, orderCreateHandler)
		ordertransport.RegisterGuestCreateAndPayRoute(guest, orderCreateHandler)
		ordertransport.RegisterGuestPreviewRoute(guest, orderPreviewHandler)
		ordertransport.RegisterGuestReadRoutes(guest, guestOrderHandler)
		paymenttransport.RegisterGuestWriteRoutes(guest, paymentWriteHandler)
		paymenttransport.RegisterGuestLatestRoute(guest, paymentLatestHandler)
	}

	// 用户认证接口
	auth := storefront.Group("/auth")
	{
		userauthtransport.RegisterUserVerifyAuthRoutes(auth, userVerifyHandler)
		userauthtransport.RegisterUserRegisterAuthRoutes(auth, userLoginHandler)
		userauthtransport.RegisterUserLoginAuthRoutes(auth, userLoginHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIPAndJSONField("email")))
		userauthtransport.RegisterUser2FAAuthRoutes(auth, user2FAHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIP))
		userauthtransport.RegisterUserTelegramAuthRoutes(auth, userTelegramHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIP))
		userauthtransport.RegisterUserTelegramOIDCAuthRoutes(auth, userTelegramOIDCHandler, RateLimitMiddleware(redisClient, loginRule, KeyByIP))
		userauthtransport.RegisterUserPasswordAuthRoutes(auth, userPasswordHandler)
	}

	// 用户接口（需鉴权）
	user := storefront.Group("")
	user.Use(UserJWTAuthMiddleware(cfg.UserJWT.SecretKey, c.UserRepo))
	{
		userauthtransport.RegisterUserProfileRoutes(user, userProfileHandler)
		auditlogtransport.RegisterUserRoutes(user, userAuditLogHandler)
		userauthtransport.RegisterUserPasswordRoutes(user, userPasswordHandler)
		userauthtransport.RegisterUserTelegramRoutes(user, userTelegramHandler)
		userauthtransport.RegisterUserTelegramOIDCRoutes(user, userTelegramOIDCHandler)
		userauthtransport.RegisterUserEmailRoutes(user, userEmailHandler)
		userauthtransport.RegisterUser2FARoutes(user, user2FAHandler)
		carttransport.RegisterUserRoutes(user, userCartHandler)
		ordertransport.RegisterUserCreateRoute(user, orderCreateHandler)
		ordertransport.RegisterUserCreateAndPayRoute(user, orderCreateHandler)
		ordertransport.RegisterUserPreviewRoute(user, orderPreviewHandler)
		ordertransport.RegisterUserPaymentChannelsRoute(user, userOrderHandler)
		ordertransport.RegisterUserReadRoutes(user, userOrderHandler)
		ordertransport.RegisterUserCancelRoute(user, userOrderHandler)
		paymenttransport.RegisterUserWriteRoutes(user, paymentWriteHandler)
		paymenttransport.RegisterUserLatestRoute(user, paymentLatestHandler)
		wallettransport.RegisterUserRoutes(user, userWalletHandler)
		giftcardtransport.RegisterUserRoutes(user, userGiftCardHandler)
		affiliatetransport.RegisterUserRoutes(user, affiliateHandler)

		resellerConsole := user.Group("/reseller")
		resellerConsole.Use(RequireMainTenantForResellerConsole())
		{
			resellertransport.RegisterUserConsoleRoutes(resellerConsole, userResellerHandler)
			resellertransport.RegisterUserProductSettingRoutes(resellerConsole, userResellerProductSettingHandler)
			resellertransport.RegisterUserFinanceRoutes(resellerConsole, userResellerFinanceHandler)
			resellertransport.RegisterUserOrderRoutes(resellerConsole, userResellerOrderHandler)
		}

		// API 对接权限（用户中心）
		apicredentialtransport.RegisterUserRoutes(user, userApiCredentialHandler)
	}
}

func registerPaymentCallbackRoutes(apiV1 *gin.RouterGroup, callbackHandler *paymentcallbacktransport.Handler, webhookHandler *paymenttransport.WebhookHandler) {
	paymentcallbacktransport.RegisterRoutes(apiV1, callbackHandler)
	paymenttransport.RegisterWebhookRoutes(apiV1, webhookHandler)
}
