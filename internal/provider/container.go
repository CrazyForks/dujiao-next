package provider

import (
	"github.com/dujiao-next/internal/authz"
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/modules/adproxy"
	"github.com/dujiao-next/internal/modules/apicredential"
	"github.com/dujiao-next/internal/modules/auditlog"
	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/modules/cardsecret"
	"github.com/dujiao-next/internal/modules/cart"
	"github.com/dujiao-next/internal/modules/catalog"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/store/gormstore"
	"github.com/dujiao-next/internal/modules/channelclient"
	"github.com/dujiao-next/internal/modules/compliance"
	"github.com/dujiao-next/internal/modules/content"
	"github.com/dujiao-next/internal/modules/coupon"
	coupongormstore "github.com/dujiao-next/internal/modules/coupon/store/gormstore"
	"github.com/dujiao-next/internal/modules/dashboard"
	"github.com/dujiao-next/internal/modules/downstreamcallback"
	giftcardgormstore "github.com/dujiao-next/internal/modules/giftcard/store/gormstore"
	"github.com/dujiao-next/internal/modules/memberlevel"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/store/gormstore"
	"github.com/dujiao-next/internal/modules/notification"
	notificationgormstore "github.com/dujiao-next/internal/modules/notification/store/gormstore"
	"github.com/dujiao-next/internal/modules/orderrisk"
	"github.com/dujiao-next/internal/modules/procurement"
	"github.com/dujiao-next/internal/modules/promotion"
	promotiongormstore "github.com/dujiao-next/internal/modules/promotion/store/gormstore"
	"github.com/dujiao-next/internal/modules/reconciliation"
	"github.com/dujiao-next/internal/modules/reseller"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	settingscontract "github.com/dujiao-next/internal/modules/settings/contract"
	"github.com/dujiao-next/internal/modules/siteconnection"
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
	AdminRepo                  repository.AdminRepository
	UserRepo                   repository.UserRepository
	UserOAuthIdentityRepo      repository.UserOAuthIdentityRepository
	EmailVerifyCodeRepo        repository.EmailVerifyCodeRepository
	OrderRepo                  repository.OrderRepository
	PaymentRepo                repository.PaymentRepository
	PaymentChannelRepo         repository.PaymentChannelRepository
	CardSecretRepo             repository.CardSecretRepository
	CardSecretBatchRepo        repository.CardSecretBatchRepository
	GiftCardRepo               *giftcardgormstore.Store
	FulfillmentRepo            repository.FulfillmentRepository
	ProductRepo                repository.ProductRepository
	ProductSKURepo             repository.ProductSKURepository
	CartRepo                   repository.CartRepository
	CouponRepo                 *coupongormstore.Store
	CouponUsageRepo            *coupongormstore.UsageStore
	PromotionRepo              *promotiongormstore.Store
	WalletRepo                 repository.WalletRepository
	OrderRefundRecordRepo      repository.OrderRefundRecordRepository
	CategoryRepo               catalog.CategoryRepository
	SettingRepo                settingscontract.Store
	UserLoginLogRepo           auditlog.UserLoginRepository
	AuthzAuditLogRepo          auditlog.AuthzRepository
	NotificationLogRepo        *notificationgormstore.LogStore
	AdminLoginLogRepo          auditlog.AdminLoginRepository
	DashboardRepo              dashboard.Repository
	AffiliateRepo              repository.AffiliateRepository
	ResellerRepo               repository.ResellerRepository
	ResellerProductSettingRepo repository.ResellerProductSettingRepository
	ResellerOperationsRepo     repository.ResellerOperationsRepository
	ApiCredentialRepo          apicredential.Repository
	SiteConnectionRepo         repository.SiteConnectionRepository
	ProductMappingRepo         *mappinggormstore.MappingStore
	SKUMappingRepo             *mappinggormstore.SKUMappingStore
	ProcurementOrderRepo       procurement.Repository
	DownstreamOrderRefRepo     repository.DownstreamOrderRefRepository
	ReconciliationJobRepo      reconciliation.JobRepository
	ReconciliationItemRepo     reconciliation.ItemRepository
	ChannelClientRepo          repository.ChannelClientRepository
	TelegramBroadcastRepo      broadcastcontract.Store
	MemberLevelRepo            memberlevel.LevelRepository
	MemberLevelPriceRepo       *memberlevelgormstore.PriceStore
	MemberLevelUserRepo        memberlevel.UserRepository

	// Services
	AuthzService                  *authz.Service
	AuthService                   *service.AuthService
	TOTPService                   *service.TOTPService
	UserTOTPService               *service.UserTOTPService
	UserAuthService               *service.UserAuthService
	TelegramAuthService           *service.TelegramAuthService
	EmailService                  *service.EmailService
	CaptchaService                *captcha.Service
	UploadService                 *upload.Service
	ProductService                *service.ProductService
	ContentPostService            *content.PostService
	ContentPostCategoryService    *content.PostCategoryService
	ContentBannerService          *content.BannerService
	ContentMediaService           *content.MediaService
	CategoryService               *catalog.CategoryService
	SettingService                *settingsapp.Service
	SitemapService                *sitemap.Service
	CartService                   *cart.Service
	WalletService                 *service.WalletService
	OrderRefundService            *service.OrderRefundService
	OrderService                  *service.OrderService
	FulfillmentService            *service.FulfillmentService
	CouponAdminService            *coupon.AdminService
	PromotionAdminService         *promotion.AdminService
	PaymentService                *service.PaymentService
	CardSecretService             *cardsecret.Service
	GiftCardService               *service.GiftCardService
	UserLoginLogService           *auditlog.UserLoginService
	AuthzAuditService             *auditlog.AuthzService
	NotificationLogService        *notification.LogService
	DashboardService              *dashboard.Service
	NotificationService           *notification.Service
	AffiliateService              *service.AffiliateService
	ResellerDomainResolver        *reseller.DomainResolver
	ResellerPricingResolver       *service.ResellerPricingResolver
	ResellerManagementService     *reseller.ManagementService
	ResellerSiteConfigService     *reseller.SiteConfigService
	ResellerProductSettingService *reseller.ProductSettingService
	ResellerAccountingService     *service.ResellerAccountingService
	ResellerOrderService          *reseller.OrderQueryService
	ResellerOperationsService     *reseller.OperationsService
	ApiCredentialService          *apicredential.Service
	SiteConnectionService         *siteconnection.Service
	ProductMappingService         *service.ProductMappingService
	ProcurementOrderService       *procurement.Service
	DownstreamCallbackService     *downstreamcallback.Service
	ReconciliationService         *reconciliation.Service
	ChannelClientService          *channelclient.Service
	TelegramBroadcastService      *broadcastapp.Service
	MemberLevelService            *memberlevel.Service
	AdProxyService                *adproxy.Service
	OrderRiskControlService       *orderrisk.Service
	ComplianceService             *compliance.Service

	PaymentProviderRegistry *paymentprovider.Registry
}
