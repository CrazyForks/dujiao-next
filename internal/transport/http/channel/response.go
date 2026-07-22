package channelhttp

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/modules/coupon"
	"github.com/dujiao-next/internal/modules/promotion"

	"github.com/gin-gonic/gin"
)

type mappedChannelError struct {
	target    error
	httpCode  int
	code      int
	errorCode string
	key       string
}

func channelErrorRule(target error, httpCode, code int, errorCode, key string) mappedChannelError {
	return mappedChannelError{target: target, httpCode: httpCode, code: code, errorCode: errorCode, key: key}
}

var channelOrderCreateErrorRules = []mappedChannelError{
	channelErrorRule(ErrRiskIPBlacklisted, http.StatusForbidden, response.CodeForbidden, "risk_blocked", "error.risk_ip_blacklisted"),
	channelErrorRule(ErrRiskEmailBlacklisted, http.StatusForbidden, response.CodeForbidden, "risk_blocked", "error.risk_email_blacklisted"),
	channelErrorRule(ErrRiskTooManyPendingOrders, http.StatusTooManyRequests, response.CodeTooManyRequests, "risk_blocked", "error.risk_too_many_pending_orders"),
	channelErrorRule(ErrRiskOrderRateLimited, http.StatusTooManyRequests, response.CodeTooManyRequests, "risk_blocked", "error.risk_order_rate_limited"),
	channelErrorRule(ErrProductSKURequired, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.order_item_invalid"),
	channelErrorRule(ErrProductSKUInvalid, http.StatusBadRequest, response.CodeBadRequest, "sku_not_found", "error.order_item_invalid"),
	channelErrorRule(ErrInvalidOrderItem, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.order_item_invalid"),
	channelErrorRule(ErrInvalidOrderAmount, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.order_amount_invalid"),
	channelErrorRule(ErrProductPurchaseNotAllowed, http.StatusBadRequest, response.CodeBadRequest, "product_unavailable", "error.product_purchase_not_allowed"),
	channelErrorRule(ErrProductMaxPurchaseExceeded, http.StatusBadRequest, response.CodeBadRequest, "quantity_limit_exceeded", "error.product_max_purchase_exceeded"),
	channelErrorRule(ErrProductMinPurchaseNotMet, http.StatusBadRequest, response.CodeBadRequest, "quantity_below_minimum", "error.product_min_purchase_not_met"),
	channelErrorRule(ErrProductNotAvailable, http.StatusBadRequest, response.CodeBadRequest, "product_unavailable", "error.product_not_available"),
	channelErrorRule(ErrManualStockInsufficient, http.StatusBadRequest, response.CodeBadRequest, "sku_out_of_stock", "error.manual_stock_insufficient"),
	channelErrorRule(ErrCardSecretInsufficient, http.StatusBadRequest, response.CodeBadRequest, "sku_out_of_stock", "error.card_secret_insufficient"),
	channelErrorRule(ErrOrderCurrencyMismatch, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.order_currency_mismatch"),
	channelErrorRule(ErrProductPriceInvalid, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.product_price_invalid"),
	channelErrorRule(coupon.ErrInvalid, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_invalid"),
	channelErrorRule(coupon.ErrNotFound, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_not_found"),
	channelErrorRule(coupon.ErrInactive, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_inactive"),
	channelErrorRule(coupon.ErrNotStarted, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_not_started"),
	channelErrorRule(coupon.ErrExpired, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_expired"),
	channelErrorRule(coupon.ErrUsageLimit, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_usage_limit"),
	channelErrorRule(coupon.ErrPerUserLimit, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_per_user_limit"),
	channelErrorRule(coupon.ErrMinAmount, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_min_amount"),
	channelErrorRule(coupon.ErrScopeInvalid, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_scope_invalid"),
	channelErrorRule(coupon.ErrPaymentRoleNotAllowed, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_payment_role_not_allowed"),
	channelErrorRule(coupon.ErrPaymentRoleGuestOnly, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_payment_role_guest_only"),
	channelErrorRule(coupon.ErrPaymentRoleMemberOnly, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_payment_role_member_only"),
	channelErrorRule(coupon.ErrMemberLevelNotAllowed, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_member_level_not_allowed"),
	channelErrorRule(coupon.ErrWholesaleDisabled, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.coupon_wholesale_disabled"),
	channelErrorRule(promotion.ErrInvalid, http.StatusBadRequest, response.CodeBadRequest, "coupon_invalid", "error.promotion_invalid"),
	channelErrorRule(ErrManualFormSchemaInvalid, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.manual_form_schema_invalid"),
	channelErrorRule(ErrManualFormRequiredMissing, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.manual_form_required_missing"),
	channelErrorRule(ErrManualFormFieldInvalid, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.manual_form_field_invalid"),
	channelErrorRule(ErrManualFormTypeInvalid, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.manual_form_type_invalid"),
	channelErrorRule(ErrManualFormOptionInvalid, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.manual_form_option_invalid"),
}

var channelPaymentCreateErrorRules = []mappedChannelError{
	channelErrorRule(ErrPaymentInvalid, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.payment_invalid"),
	channelErrorRule(ErrOrderNotFound, http.StatusNotFound, response.CodeNotFound, "order_not_found", "error.order_not_found"),
	channelErrorRule(ErrOrderStatusInvalid, http.StatusBadRequest, response.CodeBadRequest, "order_status_invalid", "error.order_status_invalid"),
	channelErrorRule(ErrPaymentChannelNotFound, http.StatusNotFound, response.CodeNotFound, "payment_method_unavailable", "error.payment_channel_not_found"),
	channelErrorRule(ErrPaymentChannelInactive, http.StatusBadRequest, response.CodeBadRequest, "payment_method_unavailable", "error.payment_channel_inactive"),
	channelErrorRule(ErrPaymentProviderUnsupported, http.StatusBadRequest, response.CodeBadRequest, "payment_method_unavailable", "error.payment_provider_not_supported"),
	channelErrorRule(ErrPaymentChannelConfigInvalid, http.StatusBadRequest, response.CodeBadRequest, "payment_method_unavailable", "error.payment_channel_config_invalid"),
	channelErrorRule(ErrPaymentGatewayRequestFailed, http.StatusBadRequest, response.CodeBadRequest, "payment_create_failed", "error.payment_gateway_request_failed"),
	channelErrorRule(ErrPaymentGatewayResponseInvalid, http.StatusBadRequest, response.CodeBadRequest, "payment_create_failed", "error.payment_gateway_response_invalid"),
	channelErrorRule(ErrPaymentCurrencyMismatch, http.StatusBadRequest, response.CodeBadRequest, "payment_create_failed", "error.payment_currency_mismatch"),
	channelErrorRule(ErrWalletOnlyPaymentRequired, http.StatusBadRequest, response.CodeBadRequest, "wallet_only_payment_required", "error.wallet_only_payment_required"),
}

func respondChannelSuccess(c *gin.Context, data interface{}) {
	shared.ChannelSuccess(c, data)
}

func respondChannelError(c *gin.Context, httpCode, code int, errorCode, key string, err error) {
	shared.ChannelError(c, httpCode, code, errorCode, key, err)
}

func respondChannelBindError(c *gin.Context, err error) { shared.ChannelBindError(c, err) }

func respondChannelMappedError(c *gin.Context, err error, rules []mappedChannelError, fallbackHTTPCode, fallbackCode int, fallbackErrorCode, fallbackKey string) {
	if seconds := retryAfter(err); seconds > 0 {
		c.Header("Retry-After", strconv.FormatInt(seconds, 10))
	}
	for _, rule := range rules {
		if errors.Is(err, rule.target) {
			respondChannelError(c, rule.httpCode, rule.code, rule.errorCode, rule.key, nil)
			return
		}
	}
	respondChannelError(c, fallbackHTTPCode, fallbackCode, fallbackErrorCode, fallbackKey, err)
}

func respondChannelOrderCreateError(c *gin.Context, err error) {
	respondChannelMappedError(c, err, channelOrderCreateErrorRules, http.StatusBadRequest, response.CodeBadRequest, "order_create_failed", "error.order_create_failed")
}

func respondChannelPaymentCreateError(c *gin.Context, err error) {
	respondChannelMappedError(c, err, channelPaymentCreateErrorRules, http.StatusBadRequest, response.CodeBadRequest, "payment_create_failed", "error.payment_create_failed")
}

func respondChannelOrderPreviewError(c *gin.Context, err error) {
	respondChannelMappedError(c, err, channelOrderCreateErrorRules, http.StatusBadRequest, response.CodeBadRequest, "order_preview_failed", "error.order_create_failed")
}

func respondChannelIdentityServiceError(c *gin.Context, err error) {
	shared.ChannelIdentityError(c, err)
}

func channelUserIDValue(primary, legacy string) string {
	return shared.ChannelUserIDValue(primary, legacy)
}

func channelUserIDFromQuery(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return channelUserIDValue(c.Query("channel_user_id"), c.Query("telegram_user_id"))
}
