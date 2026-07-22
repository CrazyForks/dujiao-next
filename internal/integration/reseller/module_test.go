package resellerintegration_test

import (
	"github.com/dujiao-next/internal/config"
	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	resellerpersistence "github.com/dujiao-next/internal/persistence/reseller"
	"github.com/dujiao-next/internal/repository"
)

// These aliases keep the integration scenarios concise while ensuring every
// service under test is constructed from the bounded reseller module.
type (
	ResellerManagementService           = resellermodule.ManagementService
	ResellerApplyInput                  = resellermodule.ResellerApplyInput
	ResellerApproveInput                = resellermodule.ResellerApproveInput
	ResellerProfileUpdateInput          = resellermodule.ResellerProfileUpdateInput
	ResellerSystemDomainInput           = resellermodule.ResellerSystemDomainInput
	ResellerProductSettingService       = resellermodule.ProductSettingService
	ResellerProductSettingInput         = resellermodule.ProductSettingInput
	ResellerProductSettingSaveInput     = resellermodule.ProductSettingSaveInput
	ResellerProductSettingUserListInput = resellermodule.ProductSettingUserListInput
	ResellerProductSettingPreviewItem   = resellermodule.ProductSettingPreviewItem
	ResellerSiteConfigInput             = resellermodule.ResellerSiteConfigInput
	ResellerAnnouncementInput           = resellermodule.ResellerAnnouncementInput
	ResellerSupportInput                = resellermodule.ResellerSupportInput
	ResellerSEOInput                    = resellermodule.ResellerSEOInput
	LocalizedTextInput                  = resellermodule.LocalizedTextInput
	ResellerSiteConfigFieldError        = resellermodule.ResellerSiteConfigFieldError
	ResellerOrderListInput              = resellermodule.OrderListInput
)

const (
	ResellerProfitStatusCredited    = resellermodule.ProfitStatusCredited
	ResellerProfitStatusPending     = resellermodule.ProfitStatusPending
	ResellerProfitStatusUnavailable = resellermodule.ProfitStatusUnavailable
)

var (
	ErrResellerApplyDisabled        = resellermodule.ErrApplyDisabled
	ErrResellerProfileInactive      = resellermodule.ErrProfileInactive
	ErrResellerProfileStatusInvalid = resellermodule.ErrProfileStatusInvalid
	ErrResellerSiteConfigInvalid    = resellermodule.ErrSiteConfigInvalid
	ErrResellerPriceBelowBase       = resellermodule.ErrPriceBelowBase
	ErrResellerMarkupExceeded       = resellermodule.ErrMarkupExceeded
	ResellerTenantContext           = resellermodule.ResellerTenantContext
)

func NewResellerManagementService(repo repository.ResellerRepository, cfg config.ResellerConfig) *resellermodule.ManagementService {
	return resellermodule.NewManagementService(resellerpersistence.NewManagementStore(repo), cfg)
}

func NewResellerProductSettingService(
	settingRepo repository.ResellerProductSettingRepository,
	resellerRepo repository.ResellerRepository,
	productRepo productcontract.Repository,
) *resellermodule.ProductSettingService {
	return resellermodule.NewProductSettingService(
		resellerpersistence.NewProductSettingStore(settingRepo, resellerRepo),
		productRepo,
	)
}

func NewResellerSiteConfigService(repo repository.ResellerRepository) *resellermodule.SiteConfigService {
	return resellermodule.NewSiteConfigService(repo)
}

func NewResellerOrderService(repo repository.ResellerRepository) *resellermodule.OrderQueryService {
	return resellermodule.NewOrderQueryService(repo)
}
