package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/provider"
	"github.com/dujiao-next/internal/payment/wechatpay"

	"github.com/shopspring/decimal"
)

func (s *PaymentService) CapturePayment(input CapturePaymentInput) (*models.Payment, error) {
	if input.PaymentID == 0 {
		return nil, ErrPaymentInvalid
	}
	payment, err := s.paymentRepo.GetByID(input.PaymentID)
	if err != nil {
		return nil, ErrPaymentUpdateFailed
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	if payment.Status == constants.PaymentStatusSuccess {
		return payment, nil
	}

	channel, err := s.channelRepo.GetByID(payment.ChannelID)
	if err != nil {
		return nil, ErrPaymentUpdateFailed
	}
	if channel == nil {
		return nil, ErrPaymentChannelNotFound
	}

	providerType := strings.ToLower(strings.TrimSpace(channel.ProviderType))
	if providerType != constants.PaymentProviderOfficial {
		return nil, ErrPaymentProviderNotSupported
	}
	if strings.TrimSpace(payment.ProviderRef) == "" {
		return nil, ErrPaymentInvalid
	}

	channelType := strings.ToLower(strings.TrimSpace(channel.ChannelType))

	// stripe + paypal 走 Registry(P1.2 Phase 1 pilot)。
	// wechat 仍走旧 switch case,P1.2b 完成 wechat adapter wrapper 后一并切。
	if channelType == constants.PaymentChannelTypeStripe || channelType == constants.PaymentChannelTypePaypal {
		return s.captureViaRegistry(input, payment, channel)
	}

	switch channelType {
	case constants.PaymentChannelTypeWechat:
		return s.captureWechatPayment(input, payment, channel)
	default:
		return nil, ErrPaymentProviderNotSupported
	}
}

func (s *PaymentService) captureWechatPayment(input CapturePaymentInput, payment *models.Payment, channel *models.PaymentChannel) (*models.Payment, error) {
	cfg, err := wechatpay.ParseConfig(channel.ConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
	}
	if err := wechatpay.ValidateConfig(cfg, channel.InteractionMode); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
	}

	ctx, cancel := detachOutboundRequestContext(input.Context)
	defer cancel()

	queryResult, err := wechatpay.QueryOrderByOutTradeNo(ctx, cfg, payment.ProviderRef)
	if err != nil {
		switch {
		case errors.Is(err, wechatpay.ErrConfigInvalid):
			return nil, fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		case errors.Is(err, wechatpay.ErrRequestFailed):
			return nil, ErrPaymentGatewayRequestFailed
		case errors.Is(err, wechatpay.ErrResponseInvalid):
			return nil, ErrPaymentGatewayResponseInvalid
		default:
			return nil, ErrPaymentGatewayRequestFailed
		}
	}

	amount := models.Money{}
	if strings.TrimSpace(queryResult.Amount) != "" {
		parsed, parseErr := decimal.NewFromString(strings.TrimSpace(queryResult.Amount))
		if parseErr == nil {
			amount = models.NewMoneyFromDecimal(parsed)
		}
	}
	payload := models.JSON{}
	if queryResult.Raw != nil {
		payload = models.JSON(queryResult.Raw)
	}
	status := strings.TrimSpace(queryResult.Status)
	if status == "" {
		status = constants.PaymentStatusPending
	}
	callbackInput := PaymentCallbackInput{
		PaymentID:   payment.ID,
		ChannelID:   channel.ID,
		Status:      status,
		ProviderRef: pickFirstNonEmpty(strings.TrimSpace(queryResult.TransactionID), strings.TrimSpace(payment.ProviderRef)),
		Amount:      amount,
		Currency:    strings.ToUpper(strings.TrimSpace(queryResult.Currency)),
		PaidAt:      queryResult.PaidAt,
		Payload:     payload,
	}
	return s.HandleCallback(callbackInput)
}

// captureViaRegistry 通过 PaymentProviderRegistry 路由调用 QueryPayment,
// 替代原 capturePaypalPayment / captureStripePayment 的内联实现。
// 仅 stripe + paypal 走此路径(P1.2 Phase 1 pilot),其它 channel 仍走旧 switch case。
func (s *PaymentService) captureViaRegistry(input CapturePaymentInput, payment *models.Payment, channel *models.PaymentChannel) (*models.Payment, error) {
	logger.Infow("payment_capture_via_registry",
		"payment_id", payment.ID,
		"provider_type", channel.ProviderType,
		"channel_type", channel.ChannelType,
	)
	if s.paymentProviderRegistry == nil {
		return nil, ErrPaymentProviderNotSupported
	}
	p, ok := s.paymentProviderRegistry.Lookup(channel.ProviderType, channel.ChannelType)
	if !ok {
		return nil, ErrPaymentProviderNotSupported
	}
	capturer, ok := p.(provider.Capturer)
	if !ok {
		logger.Warnw("payment_provider_capture_not_implemented",
			"provider_type", channel.ProviderType,
			"channel_type", channel.ChannelType,
		)
		return nil, ErrPaymentProviderNotSupported
	}

	if err := capturer.ValidateConfig(channel.ConfigJSON, channel.ChannelType); err != nil {
		return nil, mapProviderErrorToService(err)
	}

	ctx, cancel := detachOutboundRequestContext(input.Context)
	defer cancel()

	queryResult, err := capturer.QueryPayment(ctx, channel.ConfigJSON, payment.ProviderRef)
	if err != nil {
		return nil, mapProviderErrorToService(err)
	}

	payload := models.JSON{}
	if queryResult.Payload != nil {
		payload = queryResult.Payload
	}
	status := strings.TrimSpace(queryResult.Status)
	if status == "" {
		status = constants.PaymentStatusPending
	}

	callbackInput := PaymentCallbackInput{
		PaymentID:   payment.ID,
		ChannelID:   channel.ID,
		Status:      status,
		ProviderRef: pickFirstNonEmpty(queryResult.ProviderRef, payment.ProviderRef),
		Amount:      queryResult.Amount,
		Currency:    strings.ToUpper(strings.TrimSpace(queryResult.Currency)),
		PaidAt:      queryResult.PaidAt,
		Payload:     payload,
	}
	return s.HandleCallback(callbackInput)
}

// mapProviderErrorToService 把 provider.ErrXxx 转换为 service 层的 ErrPaymentXxx。
func mapProviderErrorToService(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, provider.ErrConfigInvalid):
		return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
	case errors.Is(err, provider.ErrRequestFailed), errors.Is(err, provider.ErrAuthFailed):
		return fmt.Errorf("%w: %v", ErrPaymentGatewayRequestFailed, err)
	case errors.Is(err, provider.ErrResponseInvalid), errors.Is(err, provider.ErrSignatureInvalid):
		return fmt.Errorf("%w: %v", ErrPaymentGatewayResponseInvalid, err)
	case errors.Is(err, provider.ErrUnsupportedChannel), errors.Is(err, provider.ErrProviderNotFound):
		return ErrPaymentProviderNotSupported
	default:
		return fmt.Errorf("%w: %v", ErrPaymentGatewayRequestFailed, err)
	}
}
