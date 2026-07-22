package provider

import (
	"github.com/dujiao-next/internal/models"
	apicredentialgormstore "github.com/dujiao-next/internal/modules/apicredential/store/gormstore"
	auditloggormstore "github.com/dujiao-next/internal/modules/auditlog/store/gormstore"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/store/gormstore"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"
	cataloggormstore "github.com/dujiao-next/internal/modules/catalog/store/gormstore"
	channelclientstore "github.com/dujiao-next/internal/modules/channelclient/infrastructure/gormstore"
	coupongormstore "github.com/dujiao-next/internal/modules/coupon/store/gormstore"
	dashboardgormstore "github.com/dujiao-next/internal/modules/dashboard/store/gormstore"
	giftcardgormstore "github.com/dujiao-next/internal/modules/giftcard/store/gormstore"
	emailverificationstore "github.com/dujiao-next/internal/modules/identity/emailverification/infrastructure/gormstore"
	externalidentitystore "github.com/dujiao-next/internal/modules/identity/externalidentity/infrastructure/gormstore"
	memberlevelgormstore "github.com/dujiao-next/internal/modules/memberlevel/store/gormstore"
	notificationgormstore "github.com/dujiao-next/internal/modules/notification/store/gormstore"
	procurementgormstore "github.com/dujiao-next/internal/modules/procurement/store/gormstore"
	promotiongormstore "github.com/dujiao-next/internal/modules/promotion/store/gormstore"
	reconciliationgormstore "github.com/dujiao-next/internal/modules/reconciliation/store/gormstore"
	settingsstore "github.com/dujiao-next/internal/modules/settings/infrastructure/gormstore"
	broadcaststore "github.com/dujiao-next/internal/modules/telegram/broadcast/infrastructure/gormstore"
	"github.com/dujiao-next/internal/repository"
)

func (c *Container) initRepositories() {
	db := models.DB
	c.AdminRepo = repository.NewAdminRepository(db)
	c.UserRepo = repository.NewUserRepository(db)
	c.ExternalIdentityStore = externalidentitystore.New(db)
	c.EmailVerificationStore = emailverificationstore.New(db)
	c.OrderRepo = repository.NewOrderRepository(db)
	c.PaymentRepo = repository.NewPaymentRepository(db)
	c.PaymentChannelRepo = repository.NewPaymentChannelRepository(db)
	c.CardSecretRepo = repository.NewCardSecretRepository(db)
	c.CardSecretBatchRepo = repository.NewCardSecretBatchRepository(db)
	c.GiftCardRepo = giftcardgormstore.New(db)
	c.FulfillmentRepo = repository.NewFulfillmentRepository(db)
	c.ProductRepo = repository.AdaptProductStore(productgormstore.NewProductStore(db))
	c.ProductSKURepo = repository.AdaptProductSKUStore(productgormstore.NewSKUStore(db))
	c.CartRepo = repository.NewCartRepository(db)
	c.CouponRepo = coupongormstore.New(db)
	c.CouponUsageRepo = coupongormstore.NewUsageStore(db)
	c.PromotionRepo = promotiongormstore.New(db)
	c.WalletRepo = repository.NewWalletRepository(db)
	c.OrderRefundRecordRepo = repository.NewOrderRefundRecordRepository(db)
	c.CategoryRepo = cataloggormstore.NewCategoryStore(db)
	c.SettingRepo = settingsstore.New(db)
	c.UserLoginLogRepo = auditloggormstore.NewUserLoginStore(db)
	c.AuthzAuditLogRepo = auditloggormstore.NewAuthzStore(db)
	c.NotificationLogRepo = notificationgormstore.NewLogStore(db)
	c.AdminLoginLogRepo = auditloggormstore.NewAdminLoginStore(db)
	c.DashboardRepo = dashboardgormstore.New(db)
	c.AffiliateRepo = repository.NewAffiliateRepository(db)
	c.ResellerRepo = repository.NewResellerRepository(db)
	c.ResellerProductSettingRepo = repository.NewResellerProductSettingRepository(db)
	c.ResellerOperationsRepo = repository.NewResellerOperationsRepository(db)
	c.ApiCredentialRepo = apicredentialgormstore.New(db)
	c.SiteConnectionRepo = repository.NewSiteConnectionRepository(db)
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
