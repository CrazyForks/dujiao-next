package service

import (
	"errors"

	affiliatemodule "github.com/dujiao-next/internal/modules/affiliate"
	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/modules/cardsecret"
	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	"github.com/dujiao-next/internal/modules/catalog/product/manualform"
	"github.com/dujiao-next/internal/modules/giftcard"
	"github.com/dujiao-next/internal/modules/notification"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/modules/telegram"
	walletmodule "github.com/dujiao-next/internal/modules/wallet"
)

var (
	ErrNotFound                            = catalogproduct.ErrNotFound
	ErrSlugExists                          = catalogproduct.ErrSlugExists
	ErrProductCategoryInvalid              = catalogproduct.ErrProductCategoryInvalid
	ErrInvalidCredentials                  = errors.New("invalid credentials")
	ErrInvalidPassword                     = errors.New("invalid password")
	ErrWeakPassword                        = errors.New("weak password")
	ErrEmailExists                         = errors.New("email exists")
	ErrEmailNotVerified                    = errors.New("email not verified")
	ErrUserDisabled                        = affiliatemodule.ErrUserDisabled
	ErrInvalidEmail                        = errors.New("invalid email")
	ErrEmailDomainNotAllowed               = errors.New("email domain not allowed")
	ErrInvalidVerifyPurpose                = errors.New("invalid verify purpose")
	ErrAgreementRequired                   = errors.New("agreement required")
	ErrVerifyCodeInvalid                   = errors.New("verify code invalid")
	ErrVerifyCodeExpired                   = errors.New("verify code expired")
	ErrVerifyCodeTooFrequent               = errors.New("verify code too frequent")
	ErrVerifyCodeAttemptsExceeded          = errors.New("verify code attempts exceeded")
	ErrEmailServiceDisabled                = errors.New("email service disabled")
	ErrEmailServiceNotConfigured           = errors.New("email service not configured")
	ErrEmailRecipientRejected              = errors.New("email recipient rejected")
	ErrTelegramAuthDisabled                = errors.New("telegram auth disabled")
	ErrTelegramAuthConfigInvalid           = settingsmodule.ErrTelegramAuthConfigInvalid
	ErrTelegramAuthPayloadInvalid          = errors.New("telegram auth payload invalid")
	ErrTelegramAuthSignatureInvalid        = errors.New("telegram auth signature invalid")
	ErrTelegramAuthExpired                 = errors.New("telegram auth expired")
	ErrTelegramAuthReplay                  = errors.New("telegram auth replay")
	ErrTelegramOIDCStateInvalid            = errors.New("telegram oidc state invalid")
	ErrTelegramOIDCTokenExchange           = errors.New("telegram oidc token exchange failed")
	ErrTelegramOIDCIDTokenInvalid          = errors.New("telegram oidc id token invalid")
	ErrUserOAuthIdentityExists             = errors.New("user oauth identity exists")
	ErrUserOAuthAlreadyBound               = errors.New("user oauth already bound")
	ErrUserOAuthNotBound                   = errors.New("user oauth not bound")
	ErrTelegramUnbindRequiresEmail         = errors.New("telegram unbind requires real email")
	ErrSMTPConfigInvalid                   = settingsmodule.ErrSMTPConfigInvalid
	ErrCaptchaConfigInvalid                = captcha.ErrConfigInvalid
	ErrCaptchaRequired                     = captcha.ErrRequired
	ErrCaptchaInvalid                      = captcha.ErrInvalid
	ErrCaptchaVerifyFailed                 = captcha.ErrVerifyFailed
	ErrProfileEmpty                        = errors.New("profile empty")
	ErrEmailChangeInvalid                  = errors.New("email change invalid")
	ErrEmailChangeExists                   = errors.New("email change exists")
	ErrProductPriceInvalid                 = catalogproduct.ErrProductPriceInvalid
	ErrProductPurchaseInvalid              = catalogproduct.ErrProductPurchaseInvalid
	ErrProductMaxPurchaseExceeded          = productdomain.ErrMaxPurchaseExceeded
	ErrProductMinPurchaseNotMet            = productdomain.ErrMinPurchaseNotMet
	ErrProductPurchaseLimitInvalid         = catalogproduct.ErrProductPurchaseLimitInvalid
	ErrProductStockDisplayInvalid          = catalogproduct.ErrProductStockDisplayInvalid
	ErrWholesalePriceInvalid               = productdomain.ErrWholesalePriceInvalid
	ErrManualStockInvalid                  = catalogproduct.ErrManualStockInvalid
	ErrManualStockInsufficient             = errors.New("manual stock insufficient")
	ErrUpstreamStockInsufficient           = catalogmapping.ErrUpstreamStockInsufficient
	ErrManualFormSchemaInvalid             = manualform.ErrSchemaInvalid
	ErrManualFormRequiredMissing           = manualform.ErrRequiredMissing
	ErrManualFormFieldInvalid              = manualform.ErrFieldInvalid
	ErrManualFormTypeInvalid               = manualform.ErrTypeInvalid
	ErrManualFormOptionInvalid             = manualform.ErrOptionInvalid
	ErrProductFetchFailed                  = errors.New("product fetch failed")
	ErrProductNotFound                     = errors.New("product not found")
	ErrProductSKURequired                  = errors.New("product sku required")
	ErrProductSKUInvalid                   = catalogproduct.ErrProductSKUInvalid
	ErrProductSKUHasCardSecretStock        = catalogproduct.ErrProductSKUHasCardSecretStock
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
	ErrFulfillmentInvalid                  = catalogproduct.ErrFulfillmentInvalid
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
	ErrGiftCardInvalid                     = giftcard.ErrInvalid
	ErrGiftCardNotFound                    = giftcard.ErrNotFound
	ErrGiftCardExpired                     = giftcard.ErrExpired
	ErrGiftCardDisabled                    = giftcard.ErrDisabled
	ErrGiftCardRedeemed                    = giftcard.ErrRedeemed
	ErrGiftCardCreateFailed                = giftcard.ErrCreateFailed
	ErrGiftCardFetchFailed                 = giftcard.ErrFetchFailed
	ErrGiftCardUpdateFailed                = giftcard.ErrUpdateFailed
	ErrGiftCardDeleteFailed                = giftcard.ErrDeleteFailed
	ErrGiftCardBatchCreateFailed           = giftcard.ErrBatchCreateFailed
	ErrQueueUnavailable                    = errors.New("queue unavailable")
	ErrAffiliateConfigInvalid              = settingsmodule.ErrAffiliateConfigInvalid
	ErrAffiliateDisabled                   = affiliatemodule.ErrDisabled
	ErrAffiliateNotOpened                  = affiliatemodule.ErrNotOpened
	ErrAffiliateCodeInvalid                = affiliatemodule.ErrCodeInvalid
	ErrAffiliateProfileStatusInvalid       = affiliatemodule.ErrProfileStatusInvalid
	ErrAffiliateWithdrawAmountInvalid      = affiliatemodule.ErrWithdrawAmountInvalid
	ErrAffiliateWithdrawChannelInvalid     = affiliatemodule.ErrWithdrawChannelInvalid
	ErrAffiliateWithdrawInsufficient       = affiliatemodule.ErrWithdrawInsufficient
	ErrAffiliateWithdrawStatusInvalid      = affiliatemodule.ErrWithdrawStatusInvalid
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
	ErrNotificationConfigInvalid           = notification.ErrConfigInvalid
	ErrNotificationSendFailed              = notification.ErrSendFailed
	ErrNotificationEventInvalid            = notification.ErrEventInvalid
	ErrTelegramBroadcastInvalid            = telegram.ErrBroadcastInvalid
	ErrTelegramBroadcastNotFound           = telegram.ErrBroadcastNotFound
	ErrTelegramBroadcastNoRecipients       = telegram.ErrBroadcastNoRecipients
	ErrTelegramBotTokenUnavailable         = telegram.ErrBotTokenUnavailable
	ErrRegistrationDisabled                = errors.New("registration disabled")
	ErrOrderEmailTemplateConfigInvalid     = settingsmodule.ErrOrderEmailTemplateConfigInvalid
	ErrPaymentChannelNotAllowedForProduct  = errors.New("payment channel not allowed for product")
	ErrPaymentChannelNotAllowedForRecharge = errors.New("payment channel not allowed for wallet recharge")
	ErrWalletOnlyPaymentRequired           = walletmodule.ErrOnlyPaymentRequired
	ErrResellerProductNotListed            = catalogproduct.ErrResellerProductNotListed
	ErrResellerPriceBelowBase              = resellermodule.ErrPriceBelowBase
	ErrResellerMarkupExceeded              = resellermodule.ErrMarkupExceeded
	ErrResellerCouponNotAllowed            = errors.New("reseller coupon not allowed")
	ErrResellerPricingModeInvalid          = resellermodule.ErrPricingModeInvalid
	ErrProductHasStock                     = catalogproduct.ErrProductHasStock
	ErrProductHasOrderRecord               = catalogproduct.ErrProductHasOrderRecord
)
