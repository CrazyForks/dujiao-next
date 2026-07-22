package service

import (
	"strings"

	"github.com/dujiao-next/internal/models"
)

func shouldUseGatewayOrderNo(channel *models.PaymentChannel) bool {
	return channel != nil
}

func buildGatewayOrderNo() string {
	return generateSerialNo("DJP")
}

func resolveGatewayOrderNo(channel *models.PaymentChannel, payment *models.Payment) string {
	if !shouldUseGatewayOrderNo(channel) {
		return ""
	}
	if payment != nil {
		if gatewayOrderNo := strings.TrimSpace(payment.GatewayOrderNo); gatewayOrderNo != "" {
			return gatewayOrderNo
		}
	}
	return buildGatewayOrderNo()
}

func resolveProviderOrderNo(businessOrderNo string, payment *models.Payment) string {
	if gatewayOrderNo := strings.TrimSpace(payment.GatewayOrderNo); gatewayOrderNo != "" {
		return gatewayOrderNo
	}
	return strings.TrimSpace(businessOrderNo)
}

func matchesBusinessOrderNo(callbackOrderNo string, businessOrderNo string, payment *models.Payment) bool {
	callbackOrderNo = strings.TrimSpace(callbackOrderNo)
	if callbackOrderNo == "" {
		return true
	}
	if callbackOrderNo == strings.TrimSpace(businessOrderNo) {
		return true
	}
	return callbackOrderNo == strings.TrimSpace(payment.GatewayOrderNo)
}

func buildPaymentReturnQuery(input CreatePaymentInput, order *models.Order, marker string, sessionID string) map[string]string {
	params := map[string]string{}

	bizType := strings.ToLower(strings.TrimSpace(input.ReturnBizType))
	businessNo := strings.TrimSpace(input.ReturnBusinessNo)
	isGuest := input.ReturnGuest

	if bizType == "" {
		bizType = "order"
	}
	if order != nil {
		if businessNo == "" {
			businessNo = strings.TrimSpace(order.OrderNo)
		}
		if !isGuest && order.UserID == 0 && bizType == "order" {
			isGuest = true
		}
	}

	if bizType != "" {
		params["biz_type"] = bizType
	}
	switch bizType {
	case "recharge":
		if businessNo != "" {
			params["recharge_no"] = businessNo
		}
	default:
		if businessNo != "" {
			params["order_no"] = businessNo
		}
		if isGuest {
			params["guest"] = "1"
		}
	}
	if marker = strings.TrimSpace(marker); marker != "" {
		params[marker] = "1"
	}
	if sessionID = strings.TrimSpace(sessionID); sessionID != "" {
		params["session_id"] = sessionID
	}
	return params
}
