package provider

import (
	telegrambroadcast "github.com/dujiao-next/internal/bootstrap/telegrambroadcast"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/apicredential"
	"github.com/dujiao-next/internal/modules/auditlog"
	channelclientapp "github.com/dujiao-next/internal/modules/channelclient/application"
	"github.com/dujiao-next/internal/modules/content"
	localfilestore "github.com/dujiao-next/internal/modules/content/filestore/local"
	contentgormstore "github.com/dujiao-next/internal/modules/content/store/gormstore"
	"github.com/dujiao-next/internal/modules/dashboard"
	"github.com/dujiao-next/internal/modules/downstreamcallback"
	"github.com/dujiao-next/internal/modules/notification"
	"github.com/dujiao-next/internal/modules/procurement"
	"github.com/dujiao-next/internal/modules/reconciliation"
	"github.com/dujiao-next/internal/modules/siteconnection"
	telegrammodule "github.com/dujiao-next/internal/modules/telegram"
	broadcastapp "github.com/dujiao-next/internal/modules/telegram/broadcast/application"
	"github.com/dujiao-next/internal/service"
)

// initIntegrationServices 装配通知、站点对接、支付、采购、渠道与 Telegram 集成。
func (c *Container) initIntegrationServices() {
	c.UserLoginLogService = auditlog.NewUserLoginService(c.UserLoginLogRepo)
	c.AuthzAuditService = auditlog.NewAuthzService(c.AuthzAuditLogRepo)
	c.NotificationLogService = notification.NewLogService(c.NotificationLogRepo)
	c.DashboardService = dashboard.NewService(c.DashboardRepo, c.SettingService)
	c.NotificationService = notification.NewService(
		c.SettingService,
		c.EmailService,
		c.QueueClient,
		c.DashboardService,
		c.NotificationLogService,
		telegrammodule.NewNotifyService(c.SettingService, c.Config.TelegramAuth),
	)
	c.ApiCredentialService = apicredential.NewService(c.ApiCredentialRepo)
	c.SiteConnectionService = siteconnection.NewService(c.SiteConnectionRepo, c.Config.App.SecretKey, "uploads")
	mediaCore := content.NewMediaService(
		contentgormstore.NewMediaStore(models.DB),
		localfilestore.New(),
		content.WarningLoggerFunc(logger.Warnw),
	)
	c.ContentMediaService = mediaCore
	productMappingService, err := service.NewProductMappingService(
		c.ProductMappingRepo,
		c.SKUMappingRepo,
		c.ProductRepo,
		c.ProductSKURepo,
		c.CategoryRepo,
		c.SiteConnectionService,
		mediaCore,
	)
	if err != nil {
		logger.Errorw("provider_init_product_mapping_failed", "error", err)
		panic(err)
	}
	c.ProductMappingService = productMappingService
	c.ProductMappingService.SetCategoryService(c.CategoryService)
	c.ProductMappingService.SetSettingService(c.SettingService)
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
		UserRepo:                  c.UserRepo,
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
