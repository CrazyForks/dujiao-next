package provider

import (
	"github.com/dujiao-next/internal/models"
	affiliategormstore "github.com/dujiao-next/internal/modules/affiliate/infrastructure/gormstore"
	apicredentialgormstore "github.com/dujiao-next/internal/modules/apicredential/store/gormstore"
	auditloggormstore "github.com/dujiao-next/internal/modules/auditlog/store/gormstore"
	categorygormstore "github.com/dujiao-next/internal/modules/catalog/category/infrastructure/gormstore"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/infrastructure/gormstore"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"
	channelclientstore "github.com/dujiao-next/internal/modules/channelclient/infrastructure/gormstore"
	coupongormstore "github.com/dujiao-next/internal/modules/coupon/infrastructure/gormstore"
	dashboardgormstore "github.com/dujiao-next/internal/modules/dashboard/store/gormstore"
	giftcardgormstore "github.com/dujiao-next/internal/modules/giftcard/infrastructure/gormstore"
	adminstore "github.com/dujiao-next/internal/modules/identity/admin/infrastructure/gormstore"
	emailverificationstore "github.com/dujiao-next/internal/modules/identity/emailverification/infrastructure/gormstore"
	externalidentitystore "github.com/dujiao-next/internal/modules/identity/externalidentity/infrastructure/gormstore"
	userstore "github.com/dujiao-next/internal/modules/identity/user/infrastructure/gormstore"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/infrastructure/gormstore"
	notificationgormstore "github.com/dujiao-next/internal/modules/notification/store/gormstore"
	procurementgormstore "github.com/dujiao-next/internal/modules/procurement/store/gormstore"
	promotiongormstore "github.com/dujiao-next/internal/modules/promotion/infrastructure/gormstore"
	reconciliationgormstore "github.com/dujiao-next/internal/modules/reconciliation/store/gormstore"
	settingsstore "github.com/dujiao-next/internal/modules/settings/infrastructure/gormstore"
	siteconnectiongormstore "github.com/dujiao-next/internal/modules/siteconnection/infrastructure/gormstore"
	broadcaststore "github.com/dujiao-next/internal/modules/telegram/broadcast/infrastructure/gormstore"
	"github.com/dujiao-next/internal/repository"
)

func (c *Container) initRepositories() {
	db := models.DB
	c.AdminStore = adminstore.New(db)
	c.UserStore = userstore.New(db)
	c.ExternalIdentityStore = externalidentitystore.New(db)
	c.EmailVerificationStore = emailverificationstore.New(db)
	c.OrderRepo = repository.NewOrderRepository(db)
	c.PaymentRepo = repository.NewPaymentRepository(db)
	c.PaymentChannelRepo = repository.NewPaymentChannelRepository(db)
	c.CardSecretRepo = repository.NewCardSecretRepository(db)
	c.CardSecretBatchRepo = repository.NewCardSecretBatchRepository(db)
	c.GiftCardRepo = giftcardgormstore.New(db)
	c.FulfillmentRepo = repository.NewFulfillmentRepository(db)
	c.ProductRepo = productgormstore.NewProductStore(db)
	c.ProductSKURepo = productgormstore.NewSKUStore(db)
	c.CartRepo = repository.NewCartRepository(db)
	c.CouponRepo = coupongormstore.New(db)
	c.CouponUsageRepo = coupongormstore.NewUsageStore(db)
	c.PromotionRepo = promotiongormstore.New(db)
	c.WalletRepo = repository.NewWalletRepository(db)
	c.OrderRefundRecordRepo = repository.NewOrderRefundRecordRepository(db)
	c.CategoryRepo = categorygormstore.NewCategoryStore(db)
	c.SettingRepo = settingsstore.New(db)
	c.UserLoginLogRepo = auditloggormstore.NewUserLoginStore(db)
	c.AuthzAuditLogRepo = auditloggormstore.NewAuthzStore(db)
	c.NotificationLogRepo = notificationgormstore.NewLogStore(db)
	c.AdminLoginLogRepo = auditloggormstore.NewAdminLoginStore(db)
	c.DashboardRepo = dashboardgormstore.New(db)
	c.AffiliateRepo = affiliategormstore.New(db)
	c.ResellerRepo = repository.NewResellerRepository(db)
	c.ResellerProductSettingRepo = repository.NewResellerProductSettingRepository(db)
	c.ResellerOperationsRepo = repository.NewResellerOperationsRepository(db)
	c.ApiCredentialRepo = apicredentialgormstore.New(db)
	c.SiteConnectionRepo = siteconnectiongormstore.New(db)
	c.ProductMappingRepo = mappinggormstore.NewMappingStore(db)
	c.SKUMappingRepo = mappinggormstore.NewSKUMappingStore(db)
	c.ProcurementOrderRepo = procurementgormstore.New(db)
	c.DownstreamOrderRefRepo = repository.NewDownstreamOrderRefRepository(db)
	reconciliationStore := reconciliationgormstore.New(db)
	c.ReconciliationJobRepo = reconciliationStore
	c.ReconciliationItemRepo = reconciliationgormstore.NewItemStore(reconciliationStore)
	c.ChannelClientStore = channelclientstore.New(db)
	c.TelegramBroadcastRepo = broadcaststore.New(db)
	c.MemberLevelRepo = memberlevelgormstore.NewLevelStore(db)
	c.MemberLevelPriceRepo = memberlevelgormstore.NewPriceStore(db)
	c.MemberLevelUserRepo = memberlevelgormstore.NewUserStore(db)
}
