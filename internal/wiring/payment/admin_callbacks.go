package paymentwiring

import (
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	notificationcontract "github.com/dujiao-next/internal/modules/notification/contract"
	walletcontract "github.com/dujiao-next/internal/modules/wallet/contract"
	walletdomain "github.com/dujiao-next/internal/modules/wallet/domain"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	"github.com/dujiao-next/internal/shared/jsonmap"
	paymenttransport "github.com/dujiao-next/internal/transport/http/payment"
	paymentcallbacktransport "github.com/dujiao-next/internal/transport/http/payment/callback"
)

type adminQueryAdapter struct {
	payments *service.PaymentService
}

func (a adminQueryAdapter) ListPayments(filter paymenttransport.AdminPaymentListFilter) ([]models.Payment, int64, error) {
	return a.payments.ListPayments(repository.PaymentListFilter{
		Page:         filter.Page,
		PageSize:     filter.PageSize,
		UserID:       filter.UserID,
		OrderID:      filter.OrderID,
		ChannelID:    filter.ChannelID,
		ProviderType: filter.ProviderType,
		ChannelType:  filter.ChannelType,
		Status:       filter.Status,
		CreatedFrom:  filter.CreatedFrom,
		CreatedTo:    filter.CreatedTo,
		SkipCount:    filter.SkipCount,
		Lightweight:  filter.Lightweight,
	})
}

func (a adminQueryAdapter) GetPayment(id uint) (*models.Payment, error) {
	payment, err := a.payments.GetPayment(id)
	return payment, mapTransportError(err)
}

type adminChannelLookupAdapter struct {
	channels repository.PaymentChannelRepository
}

func (a adminChannelLookupAdapter) ListByIDs(ids []uint) ([]models.PaymentChannel, error) {
	return a.channels.ListByIDs(ids)
}

type adminOrderLookupAdapter struct {
	orders repository.OrderRepository
}

func (a adminOrderLookupAdapter) GetByIDs(ids []uint) ([]models.Order, error) {
	return a.orders.GetByIDs(ids)
}

type adminRechargeLookupAdapter struct {
	wallets walletcontract.Repository
}

func (a adminRechargeLookupAdapter) GetRechargeOrdersByPaymentIDs(paymentIDs []uint) ([]walletdomain.RechargeOrder, error) {
	if a.wallets == nil {
		return nil, nil
	}
	return a.wallets.GetRechargeOrdersByPaymentIDs(paymentIDs)
}

type adminChannelCatalogAdapter struct {
	payments *service.PaymentService
	channels repository.PaymentChannelRepository
}

func (a adminChannelCatalogAdapter) ValidateChannel(channel *models.PaymentChannel) error {
	return mapTransportError(a.payments.ValidateChannel(channel))
}

func (a adminChannelCatalogAdapter) GetChannel(id uint) (*models.PaymentChannel, error) {
	channel, err := a.payments.GetChannel(id)
	return channel, mapTransportError(err)
}

func (a adminChannelCatalogAdapter) ListChannels(filter paymenttransport.AdminChannelListFilter) ([]models.PaymentChannel, int64, error) {
	return a.payments.ListChannels(repository.PaymentChannelListFilter{
		Page:         filter.Page,
		PageSize:     filter.PageSize,
		ProviderType: filter.ProviderType,
		ChannelType:  filter.ChannelType,
		ActiveOnly:   filter.ActiveOnly,
	})
}

func (a adminChannelCatalogAdapter) Create(channel *models.PaymentChannel) error {
	return a.channels.Create(channel)
}

func (a adminChannelCatalogAdapter) Update(channel *models.PaymentChannel) error {
	return a.channels.Update(channel)
}

func (a adminChannelCatalogAdapter) Delete(id uint) error {
	return a.channels.Delete(id)
}

type webhookServiceAdapter struct {
	payments *service.PaymentService
}

type callbackServiceAdapter struct {
	payments *service.PaymentService
}

func (a callbackServiceAdapter) HandleSyncCallback(channel *models.PaymentChannel, form map[string][]string, body []byte) (*models.Payment, error) {
	return a.payments.HandleSyncCallback(channel, form, body)
}

func (a callbackServiceAdapter) HandleWechatWebhook(input paymentcallbacktransport.WechatWebhookInput) (*models.Payment, string, error) {
	payment, eventType, err := a.payments.HandleWechatWebhook(service.WebhookCallbackInput{
		ChannelID: input.ChannelID,
		Headers:   input.Headers,
		Body:      input.Body,
		Context:   input.Context,
	})
	return payment, eventType, mapTransportError(err)
}

func (a webhookServiceAdapter) HandlePaypalWebhook(input paymenttransport.WebhookCallbackInput) (*models.Payment, string, error) {
	payment, eventType, err := a.payments.HandlePaypalWebhook(service.WebhookCallbackInput{
		ChannelID: input.ChannelID,
		Headers:   input.Headers,
		Body:      input.Body,
		Context:   input.Context,
	})
	return payment, eventType, mapTransportError(err)
}

func (a webhookServiceAdapter) HandleStripeWebhook(input paymenttransport.WebhookCallbackInput) (*models.Payment, string, error) {
	payment, eventType, err := a.payments.HandleStripeWebhook(service.WebhookCallbackInput{
		ChannelID: input.ChannelID,
		Headers:   input.Headers,
		Body:      input.Body,
		Context:   input.Context,
	})
	return payment, eventType, mapTransportError(err)
}

func (a webhookServiceAdapter) HandleDujiaoPayWebhook(input paymenttransport.WebhookCallbackInput) (*models.Payment, string, error) {
	payment, eventType, err := a.payments.HandleDujiaoPayWebhook(service.WebhookCallbackInput{
		ChannelID: input.ChannelID,
		Headers:   input.Headers,
		Body:      input.Body,
		Context:   input.Context,
	})
	return payment, eventType, mapTransportError(err)
}

type exceptionAlerterAdapter struct {
	notifications notificationcontract.NotificationEnqueuer
}

func (a exceptionAlerterAdapter) EnqueuePaymentExceptionAlert(method, path, clientIP string, data jsonmap.JSON) error {
	if a.notifications == nil {
		return nil
	}
	payload := jsonmap.JSON{
		"source":      constants.NotificationBizTypePaymentCallback,
		"method":      method,
		"path":        path,
		"client_ip":   clientIP,
		"occurred_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	for key, value := range data {
		payload[key] = value
	}
	return a.notifications.Enqueue(notificationcontract.EnqueueInput{
		EventType: constants.NotificationEventExceptionAlert,
		BizType:   constants.NotificationBizTypePaymentCallback,
		BizID:     0,
		Data:      payload,
	})
}
