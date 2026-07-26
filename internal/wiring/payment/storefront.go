package paymentwiring

import (
	"errors"
	"fmt"
	"time"

	orderapp "github.com/dujiao-next/internal/modules/order/application"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"

	"github.com/dujiao-next/internal/models"
	reseller "github.com/dujiao-next/internal/modules/reseller/contract"
	walletcontract "github.com/dujiao-next/internal/modules/wallet/contract"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	paymenttransport "github.com/dujiao-next/internal/transport/http/payment"
)

type guestOrderLookupAdapter struct {
	orders *orderapp.OrderService
}

func (a guestOrderLookupAdapter) GetOrderByGuestOrderNoForTenant(tenant reseller.TenantContext, orderNo, email, password string) (*orderdomain.Order, error) {
	order, err := a.orders.GetOrderByGuestOrderNoForTenant(tenant, orderNo, email, password)
	return order, mapTransportError(err)
}

func (a guestOrderLookupAdapter) GetOrderByGuestForTenant(tenant reseller.TenantContext, orderID uint, email, password string) (*orderdomain.Order, error) {
	order, err := a.orders.GetOrderByGuestForTenant(tenant, orderID, email, password)
	return order, mapTransportError(err)
}

type userOrderLookupAdapter struct {
	orders *orderapp.OrderService
}

func (a userOrderLookupAdapter) GetOrderByUserOrderNoForTenant(tenant reseller.TenantContext, orderNo string, userID uint) (*orderdomain.Order, error) {
	order, err := a.orders.GetOrderByUserOrderNoForTenant(tenant, orderNo, userID)
	return order, mapTransportError(err)
}

func (a userOrderLookupAdapter) GetOrderByUserForTenant(tenant reseller.TenantContext, orderID, userID uint) (*orderdomain.Order, error) {
	order, err := a.orders.GetOrderByUserForTenant(tenant, orderID, userID)
	return order, mapTransportError(err)
}

type pendingLookupAdapter struct {
	payments repository.PaymentRepository
}

func (a pendingLookupAdapter) GetLatestPendingByOrder(orderID uint, now time.Time) (*models.Payment, error) {
	return a.payments.GetLatestPendingByOrder(orderID, now)
}

type writerAdapter struct {
	payments *service.PaymentService
}

func (a writerAdapter) CreatePayment(input paymenttransport.CreatePaymentInput) (*paymenttransport.CreatePaymentResult, error) {
	result, err := a.payments.CreatePayment(service.CreatePaymentInput{
		OrderID:       input.OrderID,
		ChannelID:     input.ChannelID,
		UseBalance:    input.UseBalance,
		ClientIP:      input.ClientIP,
		Context:       input.Context,
		RequestScheme: input.RequestScheme,
	})
	if err != nil {
		return nil, mapTransportError(err)
	}
	if result == nil {
		return nil, nil
	}
	return &paymenttransport.CreatePaymentResult{
		Payment:          result.Payment,
		Channel:          result.Channel,
		OrderPaid:        result.OrderPaid,
		WalletPaidAmount: result.WalletPaidAmount,
		OnlinePayAmount:  result.OnlinePayAmount,
	}, nil
}

func (a writerAdapter) GetPayment(id uint) (*models.Payment, error) {
	payment, err := a.payments.GetPayment(id)
	return payment, mapTransportError(err)
}

func (a writerAdapter) CapturePayment(input paymenttransport.CapturePaymentInput) (*models.Payment, error) {
	payment, err := a.payments.CapturePayment(service.CapturePaymentInput{
		PaymentID: input.PaymentID,
		Context:   input.Context,
	})
	return payment, mapTransportError(err)
}

func mapTransportError(err error) error {
	if err == nil {
		return nil
	}
	for _, mapping := range []struct {
		source error
		target error
	}{
		{orderapp.ErrOrderNotFound, paymenttransport.ErrOrderNotFound},
		{orderapp.ErrGuestOrderNotFound, paymenttransport.ErrGuestOrderNotFound},
		{orderapp.ErrOrderStatusInvalid, paymenttransport.ErrOrderStatusInvalid},
		{service.ErrPaymentInvalid, paymenttransport.ErrPaymentInvalid},
		{service.ErrPaymentNotFound, paymenttransport.ErrPaymentNotFound},
		{service.ErrPaymentChannelNotFound, paymenttransport.ErrPaymentChannelNotFound},
		{service.ErrPaymentChannelInactive, paymenttransport.ErrPaymentChannelInactive},
		{service.ErrPaymentProviderNotSupported, paymenttransport.ErrPaymentProviderNotSupported},
		{service.ErrPaymentChannelConfigInvalid, paymenttransport.ErrPaymentChannelConfigInvalid},
		{service.ErrPaymentGatewayRequestFailed, paymenttransport.ErrPaymentGatewayRequestFailed},
		{service.ErrPaymentGatewayResponseInvalid, paymenttransport.ErrPaymentGatewayResponseInvalid},
		{service.ErrPaymentCurrencyMismatch, paymenttransport.ErrPaymentCurrencyMismatch},
		{walletcontract.ErrNotSupportedForGuest, paymenttransport.ErrWalletNotSupportedForGuest},
		{service.ErrPaymentChannelNotAllowedForProduct, paymenttransport.ErrPaymentChannelNotAllowedForProduct},
		{service.ErrPaymentChannelNotAllowedForRecharge, paymenttransport.ErrPaymentChannelNotAllowedForRecharge},
		{walletcontract.ErrOnlyPaymentRequired, paymenttransport.ErrWalletOnlyPaymentRequired},
		{service.ErrPaymentStatusInvalid, paymenttransport.ErrPaymentStatusInvalid},
		{service.ErrPaymentAmountMismatch, paymenttransport.ErrPaymentAmountMismatch},
	} {
		if errors.Is(err, mapping.source) {
			return fmt.Errorf("%w: %v", mapping.target, err)
		}
	}
	return err
}
