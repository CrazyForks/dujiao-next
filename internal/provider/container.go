package provider

import (
	"github.com/dujiao-next/internal/authz"
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/modules/adproxy"
	affiliateapp "github.com/dujiao-next/internal/modules/affiliate/application"
	affiliatecontract "github.com/dujiao-next/internal/modules/affiliate/contract"
	affiliategormstore "github.com/dujiao-next/internal/modules/affiliate/infrastructure/gormstore"
	apicredentialapp "github.com/dujiao-next/internal/modules/apicredential/application"
	apicredentialcontract "github.com/dujiao-next/internal/modules/apicredential/contract"
	auditlogapp "github.com/dujiao-next/internal/modules/auditlog/application"
	auditlogcontract "github.com/dujiao-next/internal/modules/auditlog/contract"
	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/modules/cardsecret"
	"github.com/dujiao-next/internal/modules/cart"
	categoryapp "github.com/dujiao-next/internal/modules/catalog/category/application"
	categorycontract "github.com/dujiao-next/internal/modules/catalog/category/contract"
	mappingapp "github.com/dujiao-next/internal/modules/catalog/mapping/application"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/infrastructure/gormstore"
	productapplication "github.com/dujiao-next/internal/modules/catalog/product/application"
	productadmin "github.com/dujiao-next/internal/modules/catalog/product/application/admin"
	productwrite "github.com/dujiao-next/internal/modules/catalog/product/application/write"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"
	channelclientapp "github.com/dujiao-next/internal/modules/channelclient/application"
	channelclientcontract "github.com/dujiao-next/internal/modules/channelclient/contract"
	"github.com/dujiao-next/internal/modules/compliance"
	"github.com/dujiao-next/internal/modules/content"
	couponapp "github.com/dujiao-next/internal/modules/coupon/application"
	coupongormstore "github.com/dujiao-next/internal/modules/coupon/infrastructure/gormstore"
	"github.com/dujiao-next/internal/modules/dashboard"
	"github.com/dujiao-next/internal/modules/downstreamcallback"
	giftcardapp "github.com/dujiao-next/internal/modules/giftcard/application"
	giftcardgormstore "github.com/dujiao-next/internal/modules/giftcard/infrastructure/gormstore"
	admincontract "github.com/dujiao-next/internal/modules/identity/admin/contract"
	adminauthapp "github.com/dujiao-next/internal/modules/identity/adminauth/application"
	admintotpapp "github.com/dujiao-next/internal/modules/identity/adminauth/totp/application"
	emailverificationcontract "github.com/dujiao-next/internal/modules/identity/emailverification/contract"
	externalidentitycontract "github.com/dujiao-next/internal/modules/identity/externalidentity/contract"
	telegramauthapp "github.com/dujiao-next/internal/modules/identity/telegramauth/application"
	usercontract "github.com/dujiao-next/internal/modules/identity/user/contract"
	userauthapp "github.com/dujiao-next/internal/modules/identity/userauth/application"
	usertotpapp "github.com/dujiao-next/internal/modules/identity/userauth/totp/application"
	memberlevelapp "github.com/dujiao-next/internal/modules/memberlevel/application"
	memberlevelcontract "github.com/dujiao-next/internal/modules/memberlevel/contract"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/infrastructure/gormstore"
	"github.com/dujiao-next/internal/modules/notification"
	notificationgormstore "github.com/dujiao-next/internal/modules/notification/store/gormstore"
	"github.com/dujiao-next/internal/modules/orderrisk"
	"github.com/dujiao-next/internal/modules/procurement"
	promotionapp "github.com/dujiao-next/internal/modules/promotion/application"
	promotiongormstore "github.com/dujiao-next/internal/modules/promotion/infrastructure/gormstore"
	"github.com/dujiao-next/internal/modules/reconciliation"
	"github.com/dujiao-next/internal/modules/reseller"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	settingscontract "github.com/dujiao-next/internal/modules/settings/contract"
	siteconnectionapp "github.com/dujiao-next/internal/modules/siteconnection/application"
	siteconnectioncontract "github.com/dujiao-next/internal/modules/siteconnection/contract"
	"github.com/dujiao-next/internal/modules/sitemap"
	broadcastapp "github.com/dujiao-next/internal/modules/telegram/broadcast/application"
	broadcastcontract "github.com/dujiao-next/internal/modules/telegram/broadcast/contract"
	"github.com/dujiao-next/internal/modules/upload"
	paymentprovider "github.com/dujiao-next/internal/payment/provider"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
)

// Container 声明应用运行期共享的依赖表面；具体构造过程按职责拆分在同包装配文件中。
type Container struct {
	Config      *config.Config
	QueueClient *queue.Client

	// Repositories
	AdminStore                 admincontract.Store
	UserStore                  usercontract.Store
	ExternalIdentityStore      externalidentitycontract.Store
	EmailVerificationStore     emailverificationcontract.Store
	OrderRepo                  repository.OrderRepository
	PaymentRepo                repository.PaymentRepository
	PaymentChannelRepo         repository.PaymentChannelRepository
	CardSecretRepo             repository.CardSecretRepository
	CardSecretBatchRepo        repository.CardSecretBatchRepository
	GiftCardRepo               *giftcardgormstore.Store
	FulfillmentRepo            repository.FulfillmentRepository
	ProductRepo                *productgormstore.ProductStore
	ProductSKURepo             *productgormstore.SKUStore
	CartRepo                   repository.CartRepository
	CouponRepo                 *coupongormstore.Store
	CouponUsageRepo            *coupongormstore.UsageStore
	PromotionRepo              *promotiongormstore.Store
	WalletRepo                 repository.WalletRepository
	OrderRefundRecordRepo      repository.OrderRefundRecordRepository
	CategoryRepo               categorycontract.Repository
	SettingRepo                settingscontract.Store
	UserLoginLogRepo           auditlogcontract.UserLoginRepository
	AuthzAuditLogRepo          auditlogcontract.AuthzRepository
	NotificationLogRepo        *notificationgormstore.LogStore
	AdminLoginLogRepo          auditlogcontract.AdminLoginRepository
	DashboardRepo              dashboard.Repository
	AffiliateRepo              affiliatecontract.Store
	ResellerRepo               repository.ResellerRepository
	ResellerProductSettingRepo repository.ResellerProductSettingRepository
	ResellerOperationsRepo     repository.ResellerOperationsRepository
	ApiCredentialRepo          apicredentialcontract.Repository
	SiteConnectionRepo         siteconnectioncontract.Repository
	ProductMappingRepo         *mappinggormstore.MappingStore
	SKUMappingRepo             *mappinggormstore.SKUMappingStore
	ProcurementOrderRepo       procurement.Repository
	DownstreamOrderRefRepo     repository.DownstreamOrderRefRepository
	ReconciliationJobRepo      reconciliation.JobRepository
	ReconciliationItemRepo     reconciliation.ItemRepository
	ChannelClientStore         channelclientcontract.Store
	TelegramBroadcastRepo      broadcastcontract.Store
	MemberLevelRepo            memberlevelcontract.LevelRepository
	MemberLevelPriceRepo       *memberlevelgormstore.PriceStore
	MemberLevelUserRepo        memberlevelcontract.UserRepository

	// Services
	AuthzService                  *authz.Service
	AuthService                   *adminauthapp.Service
	TOTPService                   *admintotpapp.Service
	UserTOTPService               *usertotpapp.Service
	UserAuthService               *userauthapp.Service
	TelegramAuthService           *telegramauthapp.Service
	EmailService                  *service.EmailService
	CaptchaService                *captcha.Service
	UploadService                 *upload.Service
	ProductReadService            *productapplication.Service
	ProductAdminService           *productadmin.AdminService
	ProductWriteService           *productwrite.WriteService
	ContentPostService            *content.PostService
	ContentPostCategoryService    *content.PostCategoryService
	ContentBannerService          *content.BannerService
	ContentMediaService           *content.MediaService
	CategoryService               *categoryapp.Service
	SettingService                *settingsapp.Service
	SitemapService                *sitemap.Service
	CartService                   *cart.Service
	WalletService                 *service.WalletService
	OrderRefundService            *service.OrderRefundService
	OrderService                  *service.OrderService
	FulfillmentService            *service.FulfillmentService
	CouponAdminService            *couponapp.AdminService
	PromotionAdminService         *promotionapp.AdminService
	PaymentService                *service.PaymentService
	CardSecretService             *cardsecret.Service
	GiftCardService               *giftcardapp.Service
	UserLoginLogService           *auditlogapp.UserLoginService
	AuthzAuditService             *auditlogapp.AuthzService
	AdminLoginLogService          *auditlogapp.AdminLoginService
	NotificationLogService        *notification.LogService
	DashboardService              *dashboard.Service
	NotificationService           *notification.Service
	AffiliateService              *affiliateapp.Service
	AffiliateRefundHandler        *affiliategormstore.RefundHandler
	ResellerDomainResolver        *reseller.DomainResolver
	ResellerPricingResolver       *service.ResellerPricingResolver
	ResellerManagementService     *reseller.ManagementService
	ResellerSiteConfigService     *reseller.SiteConfigService
	ResellerProductSettingService *reseller.ProductSettingService
	ResellerAccountingService     *service.ResellerAccountingService
	ResellerOrderService          *reseller.OrderQueryService
	ResellerOperationsService     *reseller.OperationsService
	ApiCredentialService          *apicredentialapp.Service
	SiteConnectionService         *siteconnectionapp.Service
	ProductMappingService         *mappingapp.Service
	ProcurementOrderService       *procurement.Service
	DownstreamCallbackService     *downstreamcallback.Service
	ReconciliationService         *reconciliation.Service
	ChannelClientService          *channelclientapp.Service
	TelegramBroadcastService      *broadcastapp.Service
	MemberLevelService            *memberlevelapp.Service
	AdProxyService                *adproxy.Service
	OrderRiskControlService       *orderrisk.Service
	ComplianceService             *compliance.Service

	PaymentProviderRegistry *paymentprovider.Registry
}
