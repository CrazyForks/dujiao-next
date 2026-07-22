package provider

import (
	catalogmappingbootstrap "github.com/dujiao-next/internal/bootstrap/catalogmapping"
	telegrambroadcast "github.com/dujiao-next/internal/bootstrap/telegrambroadcast"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	apicredentialapp "github.com/dujiao-next/internal/modules/apicredential/application"
	auditlogapp "github.com/dujiao-next/internal/modules/auditlog/application"
	channelclientapp "github.com/dujiao-next/internal/modules/channelclient/application"
	contentapp "github.com/dujiao-next/internal/modules/content/application"
	localfilestore "github.com/dujiao-next/internal/modules/content/infrastructure/filestore/local"
	contentgormstore "github.com/dujiao-next/internal/modules/content/infrastructure/gormstore"
	dashboardapp "github.com/dujiao-next/internal/modules/dashboard/application"
	"github.com/dujiao-next/internal/modules/downstreamcallback"
	notificationapp "github.com/dujiao-next/internal/modules/notification/application"
	notificationasyncqueue "github.com/dujiao-next/internal/modules/notification/infrastructure/asyncqueue"
	"github.com/dujiao-next/internal/modules/procurement"
	"github.com/dujiao-next/internal/modules/reconciliation"
	siteconnectionapp "github.com/dujiao-next/internal/modules/siteconnection/application"
	telegrammodule "github.com/dujiao-next/internal/modules/telegram"
	broadcastapp "github.com/dujiao-next/internal/modules/telegram/broadcast/application"
	"github.com/dujiao-next/internal/service"
)

// initIntegrationServices 装配通知、站点对接、支付、采购、渠道与 Telegram 集成。
func (c *Container) initIntegrationServices() {
	c.UserLoginLogService = auditlogapp.NewUserLoginService(c.UserLoginLogRepo)
	c.AuthzAuditService = auditlogapp.NewAuthzService(c.AuthzAuditLogRepo)
	c.AdminLoginLogService = auditlogapp.NewAdminLoginService(c.AdminLoginLogRepo)
	c.NotificationLogService = notificationapp.NewLogService(c.NotificationLogRepo)
	c.DashboardService = dashboardapp.NewService(c.DashboardRepo, c.SettingService)
	c.NotificationService = notificationapp.NewService(
		c.SettingService,
		c.EmailService,
		notificationasyncqueue.New(c.QueueClient),
		c.DashboardService,
		c.NotificationLogService,
		telegrammodule.NewNotifyService(c.SettingService, c.Config.TelegramAuth),
	)
	c.ApiCredentialService = apicredentialapp.NewService(c.ApiCredentialRepo)
	c.SiteConnectionService = siteconnectionapp.NewService(c.SiteConnectionRepo, c.Config.App.SecretKey, "uploads")
	mediaCore := contentapp.NewMediaService(
		contentgormstore.NewMediaStore(models.DB),
		localfilestore.New(),
		contentapp.WarningLoggerFunc(logger.Warnw),
	)
	c.ContentMediaService = mediaCore
	productMappingService, err := catalogmappingbootstrap.New(catalogmappingbootstrap.Dependencies{
		Mappings:    c.ProductMappingRepo,
		SKUMappings: c.SKUMappingRepo,
		Products:    c.ProductRepo,
		SKUs:        c.ProductSKURepo,
		Categories:  c.CategoryRepo,
		Connections: c.SiteConnectionService,
		Media:       mediaCore,
	})
	if err != nil {
		logger.Errorw("provider_init_product_mapping_failed", "error", err)
		panic(err)
	}
	c.ProductMappingService = productMappingService
	c.ProductMappingService.SetCategoryCreator(c.CategoryService)
	c.ProductMappingService.SetSettings(c.SettingService)
	c.SiteConnectionService.SetMarkupReapplier(c.ProductMappingService)
	c.OrderService.SetProductMappingService(c.ProductMappingService)
	c.DownstreamCallbackService = downstreamcallback.NewService(c.DownstreamOrderRefRepo, c.OrderRepo, c.ApiCredentialRepo, c.QueueClient)
	c.PaymentService = service.NewPaymentService(service.PaymentServiceOptions{
		OrderRepo:                 c.OrderRepo,
		ProductRepo:               c.ProductRepo,
		ProductSKURepo:            c.ProductSKURepo,
		PaymentRepo:               c.PaymentRepo,
		ChannelRepo:               c.PaymentChannelRepo,
		WalletRepo:                c.WalletRepo,
		UserStore:                 c.UserStore,
		ExternalIdentityStore:     c.ExternalIdentityStore,
		QueueClient:               c.QueueClient,
		WalletService:             c.WalletService,
		SettingService:            c.SettingService,
		DefaultEmailConfig:        c.Config.Email,
		ExpireMinutes:             c.Config.Order.PaymentExpireMinutes,
		AffiliateService:          c.AffiliateService,
		NotificationService:       c.NotificationService,
		PaymentProviderRegistry:   c.PaymentProviderRegistry,
		ResellerAccountingService: c.ResellerAccountingService,
	})
	c.ProcurementOrderService = procurement.NewService(procurement.ServiceOptions{
		Repository:         c.ProcurementOrderRepo,
		Orders:             c.OrderRepo,
		ProductMappings:    c.ProductMappingRepo,
		SKUMappings:        c.SKUMappingRepo,
		Connections:        c.SiteConnectionService,
		Queue:              c.QueueClient,
		OrderLifecycle:     service.NewProcurementOrderLifecycle(c.OrderRepo, c.FulfillmentRepo, c.QueueClient, c.SettingService, c.Config.Email),
		DownstreamCallback: c.DownstreamCallbackService,
		BotNotifier:        c.FulfillmentService,
		Notifications:      c.NotificationService,
	})
	c.ReconciliationService = reconciliation.NewService(reconciliation.ServiceOptions{
		Jobs: c.ReconciliationJobRepo, Items: c.ReconciliationItemRepo,
		Procurements: c.ProcurementOrderRepo, Connections: c.SiteConnectionService,
		Queue: c.QueueClient, Notifications: c.NotificationService,
	})
	c.ChannelClientService = channelclientapp.NewService(c.ChannelClientStore, c.Config.App.SecretKey)
	c.TelegramBroadcastService = broadcastapp.NewService(
		c.TelegramBroadcastRepo,
		telegrambroadcast.NewUserDirectory(c.ExternalIdentityStore),
		telegrambroadcast.NewBotTokenResolver(c.ChannelClientService),
		telegrambroadcast.NewDispatcher(c.QueueClient),
		telegrammodule.NewNotifyService(c.SettingService, c.Config.TelegramAuth),
	)
}
