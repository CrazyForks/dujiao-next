package provider

import (
	"github.com/dujiao-next/internal/authz"
	telegramauthcache "github.com/dujiao-next/internal/bootstrap/telegramauthcache"
	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/modules/compliance"
	telegramauthapp "github.com/dujiao-next/internal/modules/identity/telegramauth/application"
	"github.com/dujiao-next/internal/modules/reseller"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	settingsmessaging "github.com/dujiao-next/internal/modules/settings/schema/messaging"
	settingssecurity "github.com/dujiao-next/internal/modules/settings/schema/security"
	"github.com/dujiao-next/internal/modules/upload"
	resellerpersistence "github.com/dujiao-next/internal/persistence/reseller"
	"github.com/dujiao-next/internal/service"
)

// initPolicyAndSettingServices 装配授权、动态设置、分销商基础能力与合规服务。
func (c *Container) initPolicyAndSettingServices() {
	authzService, err := authz.NewService(models.DB)
	if err != nil {
		logger.Errorw("provider_init_authz_failed", "error", err)
		panic(err)
	}
	c.AuthzService = authzService
	if err := c.AuthzService.BootstrapBuiltinRoles(); err != nil {
		logger.Errorw("provider_bootstrap_builtin_roles_failed", "error", err)
		panic(err)
	}

	c.SettingService = settingsapp.NewService(c.SettingRepo, c.Config.Order)
	c.ResellerDomainResolver = reseller.NewDomainResolver(c.ResellerRepo, c.Config.Reseller)
	c.ResellerPricingResolver = service.NewResellerPricingResolver(c.ResellerRepo)
	c.ResellerManagementService = reseller.NewManagementService(resellerpersistence.NewManagementStore(c.ResellerRepo), c.Config.Reseller)
	c.ResellerSiteConfigService = reseller.NewSiteConfigService(c.ResellerRepo)
	c.ResellerProductSettingService = reseller.NewProductSettingService(
		resellerpersistence.NewProductSettingStore(c.ResellerProductSettingRepo, c.ResellerRepo),
		c.ProductRepo,
	)
	c.ResellerAccountingService = service.NewResellerAccountingService(c.ResellerRepo, service.ResellerAccountingOptions{
		ConfirmDays: c.Config.Reseller.SettlementConfirmDays,
	})
	c.ResellerOrderService = reseller.NewOrderQueryService(c.ResellerRepo)
	c.ResellerOperationsService = reseller.NewOperationsService(c.ResellerOperationsRepo)
	c.ComplianceService = compliance.NewService(c.SettingRepo)
}

// loadRuntimeSettings 用数据库设置覆盖启动配置中的可动态配置项。
func (c *Container) loadRuntimeSettings() {
	smtpSetting, err := c.SettingService.GetSMTPSetting(c.Config.Email)
	if err != nil {
		logger.Warnw("provider_load_smtp_setting_failed", "error", err)
	} else {
		c.Config.Email = settingsmessaging.SMTPSettingToConfig(smtpSetting)
	}

	captchaSetting, err := c.SettingService.GetCaptchaSetting(c.Config.Captcha)
	if err != nil {
		logger.Warnw("provider_load_captcha_setting_failed", "error", err)
	} else {
		c.Config.Captcha = settingssecurity.CaptchaSettingToConfig(captchaSetting)
	}

	telegramAuthSetting, err := c.SettingService.GetTelegramAuthSetting(c.Config.TelegramAuth)
	if err != nil {
		logger.Warnw("provider_load_telegram_auth_setting_failed", "error", err)
	} else {
		c.Config.TelegramAuth = settingssecurity.TelegramAuthSettingToConfig(telegramAuthSetting)
	}
}

// initIdentityAndCatalogServices 装配身份认证、上传、推广与商品读取能力。
func (c *Container) initIdentityAndCatalogServices() {
	c.EmailService = service.NewEmailService(&c.Config.Email)
	c.CaptchaService = captcha.NewService(c.SettingService, c.Config.Captcha)
	c.AuthService = service.NewAuthService(c.Config, c.AdminStore)
	c.TOTPService = service.NewTOTPService(c.Config, c.AdminStore, cache.Client())
	c.UserTOTPService = service.NewUserTOTPService(c.Config, c.UserStore, cache.Client())
	c.TelegramAuthService = telegramauthapp.NewService(c.Config.TelegramAuth, telegramauthcache.Options()...)
	c.UserAuthService = service.NewUserAuthService(c.Config, c.UserStore, c.ExternalIdentityStore, c.EmailVerificationStore, c.SettingService, c.EmailService, c.TelegramAuthService)
	c.UploadService = upload.NewService(c.Config)
	c.AffiliateService = service.NewAffiliateService(c.AffiliateRepo, c.UserStore, c.OrderRepo, c.ProductRepo, c.SettingService)
	c.ProductService = service.NewProductService(c.ProductRepo, c.ProductSKURepo, c.CardSecretRepo, c.CardSecretBatchRepo, c.CategoryRepo, c.MemberLevelPriceRepo, c.CartRepo, c.ProductMappingRepo, c.OrderRepo, c.PaymentChannelRepo)
}
