package service

import (
	"errors"

	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"

	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/modules/cardsecret"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	"github.com/dujiao-next/internal/modules/catalog/product/manualform"
	notificationcontract "github.com/dujiao-next/internal/modules/notification/contract"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	settingsintegration "github.com/dujiao-next/internal/modules/settings/schema/integration"
	settingsmessaging "github.com/dujiao-next/internal/modules/settings/schema/messaging"
	walletmodule "github.com/dujiao-next/internal/modules/wallet"
)

var (
	ErrNotFound                            = productcontract.ErrNotFound
	ErrSlugExists                          = productcontract.ErrSlugExists
	ErrProductCategoryInvalid              = productcontract.ErrProductCategoryInvalid
	ErrInvalidEmail                        = errors.New("invalid email")
	ErrEmailServiceDisabled                = errors.New("email service disabled")
	ErrEmailServiceNotConfigured           = errors.New("email service not configured")
	ErrEmailRecipientRejected              = errors.New("email recipient rejected")
	ErrSMTPConfigInvalid                   = settingsmessaging.ErrSMTPConfigInvalid
	ErrCaptchaConfigInvalid                = captcha.ErrConfigInvalid
	ErrCaptchaRequired                     = captcha.ErrRequired
	ErrCaptchaInvalid                      = captcha.ErrInvalid
	ErrCaptchaVerifyFailed                 = captcha.ErrVerifyFailed
	ErrProductPriceInvalid                 = productcontract.ErrProductPriceInvalid
	ErrProductPurchaseInvalid              = productcontract.ErrProductPurchaseInvalid
	ErrProductMaxPurchaseExceeded          = productdomain.ErrMaxPurchaseExceeded
	ErrProductMinPurchaseNotMet            = productdomain.ErrMinPurchaseNotMet
	ErrProductPurchaseLimitInvalid         = productcontract.ErrProductPurchaseLimitInvalid
	ErrProductStockDisplayInvalid          = productcontract.ErrProductStockDisplayInvalid
	ErrWholesalePriceInvalid               = productdomain.ErrWholesalePriceInvalid
	ErrManualStockInvalid                  = productcontract.ErrManualStockInvalid
	ErrManualStockInsufficient             = errors.New("manual stock insufficient")
	ErrManualFormSchemaInvalid             = manualform.ErrSchemaInvalid
	ErrManualFormRequiredMissing           = manualform.ErrRequiredMissing
	ErrManualFormFieldInvalid              = manualform.ErrFieldInvalid
	ErrManualFormTypeInvalid               = manualform.ErrTypeInvalid
	ErrManualFormOptionInvalid             = manualform.ErrOptionInvalid
	ErrProductFetchFailed                  = errors.New("product fetch failed")
	ErrProductNotFound                     = errors.New("product not found")
	ErrProductSKURequired                  = errors.New("product sku required")
	ErrProductSKUInvalid                   = productcontract.ErrProductSKUInvalid
	ErrProductSKUHasCardSecretStock        = productcontract.ErrProductSKUHasCardSecretStock
	ErrInvalidOrderItem                    = productdomain.ErrPurchaseQuantityInvalid
	ErrInvalidOrderAmount                  = errors.New("invalid order amount")
	ErrOrderCurrencyMismatch               = errors.New("order currency mismatch")
	ErrOrderNotFound                       = resellermodule.ErrOrderNotFound
	ErrOrderCreateFailed                   = errors.New("order create failed")
	ErrOrderFetchFailed                    = errors.New("order fetch failed")
	ErrProductNotAvailable                 = errors.New("product not available")
	ErrProductPurchaseNotAllowed           = errors.New("product purchase not allowed")
	ErrOrderStatusInvalid                  = errors.New("order status invalid")
	ErrOrderRefundExpired                  = errors.New("order refund expired")
	ErrOrderCancelNotAllowed               = errors.New("order cancel not allowed")
	ErrOrderUpdateFailed                   = errors.New("order update failed")
	ErrGuestOrderNotFound                  = errors.New("guest order not found")
	ErrGuestEmailRequired                  = errors.New("guest email required")
	ErrGuestPasswordRequired               = errors.New("guest password required")
	ErrGuestCouponNotAllowed               = errors.New("guest coupon not allowed")
	ErrFulfillmentInvalid                  = productcontract.ErrFulfillmentInvalid
	ErrFulfillmentExists                   = errors.New("fulfillment exists")
	ErrFulfillmentCreateFailed             = errors.New("fulfillment create failed")
	ErrPaymentInvalid                      = errors.New("payment invalid")
	ErrPaymentNotFound                     = errors.New("payment not found")
	ErrPaymentCreateFailed                 = errors.New("payment create failed")
	ErrPaymentUpdateFailed                 = errors.New("payment update failed")
	ErrPaymentStatusInvalid                = errors.New("payment status invalid")
	ErrPaymentAmountMismatch               = errors.New("payment amount mismatch")
	ErrPaymentCurrencyMismatch             = errors.New("payment currency mismatch")
	ErrPaymentChannelNotFound              = errors.New("payment channel not found")
	ErrPaymentChannelInactive              = errors.New("payment channel inactive")
	ErrPaymentProviderNotSupported         = errors.New("payment provider not supported")
	ErrPaymentChannelConfigInvalid         = errors.New("payment channel config invalid")
	ErrPaymentAmountTooSmall               = errors.New("payment amount too small")
	ErrPaymentAmountTooLarge               = errors.New("payment amount too large")
	ErrPaymentGatewayRequestFailed         = errors.New("payment gateway request failed")
	ErrPaymentGatewayResponseInvalid       = errors.New("payment gateway response invalid")
	ErrWalletInvalidAmount                 = walletmodule.ErrInvalidAmount
	ErrWalletInsufficientBalance           = walletmodule.ErrInsufficientBalance
	ErrWalletAccountNotFound               = walletmodule.ErrAccountNotFound
	ErrWalletAccountCreateFailed           = walletmodule.ErrAccountCreateFailed
	ErrWalletAccountUpdateFailed           = walletmodule.ErrAccountUpdateFailed
	ErrWalletTransactionCreateFailed       = walletmodule.ErrTransactionCreateFailed
	ErrWalletRefundExceeded                = walletmodule.ErrRefundExceeded
	ErrWalletNotSupportedForGuest          = walletmodule.ErrNotSupportedForGuest
	ErrWalletRechargeNotFound              = walletmodule.ErrRechargeNotFound
	ErrWalletRechargeStatusInvalid         = walletmodule.ErrRechargeStatusInvalid
	ErrRefundRecordCreateFailed            = errors.New("refund record create failed")
	ErrCardSecretInsufficient              = cardsecret.ErrInsufficient
	ErrFulfillmentNotAuto                  = errors.New("fulfillment not auto")
	ErrCardSecretInvalid                   = cardsecret.ErrInvalid
	ErrCardSecretCreateFailed              = cardsecret.ErrCreateFailed
	ErrCardSecretFetchFailed               = cardsecret.ErrFetchFailed
	ErrCardSecretUpdateFailed              = cardsecret.ErrUpdateFailed
	ErrCardSecretDeleteFailed              = cardsecret.ErrDeleteFailed
	ErrCardSecretBatchCreateFailed         = cardsecret.ErrBatchCreateFailed
	ErrCardSecretBatchFetchFailed          = cardsecret.ErrBatchFetchFailed
	ErrCardSecretImportFailed              = cardsecret.ErrImportFailed
	ErrCardSecretStatsFailed               = cardsecret.ErrStatsFailed
	ErrQueueUnavailable                    = errors.New("queue unavailable")
	ErrAffiliateConfigInvalid              = settingsintegration.ErrAffiliateConfigInvalid
	ErrResellerAccountingUnavailable       = resellermodule.ErrAccountingUnavailable
	ErrResellerLedgerInvalidSnapshot       = resellermodule.ErrLedgerInvalidSnapshot
	ErrResellerDisabled                    = errors.New("reseller disabled")
	ErrResellerApplyDisabled               = resellermodule.ErrApplyDisabled
	ErrResellerNotOpened                   = resellermodule.ErrNotOpened
	ErrResellerProfileInactive             = resellermodule.ErrProfileInactive
	ErrResellerProfileStatusInvalid        = resellermodule.ErrProfileStatusInvalid
	ErrResellerDomainInvalid               = resellermodule.ErrDomainInvalid
	ErrResellerDomainConflict              = resellermodule.ErrDomainConflict
	ErrResellerDomainStatusInvalid         = resellermodule.ErrDomainStatusInvalid
	ErrResellerSubdomainBaseMissing        = resellermodule.ErrSubdomainBaseMissing
	ErrResellerDomainMainHostNotAllowed    = resellermodule.ErrDomainMainHostNotAllowed
	ErrResellerSiteConfigInvalid           = resellermodule.ErrSiteConfigInvalid
	ErrResellerSiteConfigNotFound          = resellermodule.ErrSiteConfigNotFound
	ErrResellerSettlementUnavailable       = resellermodule.ErrSettlementUnavailable
	ErrResellerWithdrawAmountInvalid       = resellermodule.ErrWithdrawAmountInvalid
	ErrResellerWithdrawInsufficient        = resellermodule.ErrWithdrawInsufficient
	ErrResellerWithdrawCurrencyUnavailable = resellermodule.ErrWithdrawCurrencyUnavailable
	ErrResellerWithdrawStatusInvalid       = resellermodule.ErrWithdrawStatusInvalid
	ErrResellerBalanceAccountFrozen        = resellermodule.ErrBalanceAccountFrozen
	ErrNotificationConfigInvalid           = notificationcontract.ErrConfigInvalid
	ErrNotificationSendFailed              = notificationcontract.ErrSendFailed
	ErrNotificationEventInvalid            = notificationcontract.ErrEventInvalid
	ErrOrderEmailTemplateConfigInvalid     = settingsmessaging.ErrOrderEmailTemplateConfigInvalid
	ErrPaymentChannelNotAllowedForProduct  = errors.New("payment channel not allowed for product")
	ErrPaymentChannelNotAllowedForRecharge = errors.New("payment channel not allowed for wallet recharge")
	ErrWalletOnlyPaymentRequired           = walletmodule.ErrOnlyPaymentRequired
	ErrResellerProductNotListed            = productcontract.ErrResellerProductNotListed
	ErrResellerPriceBelowBase              = resellermodule.ErrPriceBelowBase
	ErrResellerMarkupExceeded              = resellermodule.ErrMarkupExceeded
	ErrResellerCouponNotAllowed            = errors.New("reseller coupon not allowed")
	ErrResellerPricingModeInvalid          = resellermodule.ErrPricingModeInvalid
	ErrProductHasStock                     = productcontract.ErrProductHasStock
	ErrProductHasOrderRecord               = productcontract.ErrProductHasOrderRecord
)
