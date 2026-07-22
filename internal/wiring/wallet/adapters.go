package walletwiring

import (
	"errors"
	"fmt"

	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	"github.com/dujiao-next/internal/shared/money"
	wallettransport "github.com/dujiao-next/internal/transport/http/wallet"
)

// walletTransportAdapter 将 legacy 钱包和支付服务适配为用户钱包 transport 端口。
type walletTransportAdapter struct {
	wallets  *service.WalletService
	payments *service.PaymentService
}

func (a walletTransportAdapter) GetAccount(userID uint) (*models.WalletAccount, error) {
	account, err := a.wallets.GetAccount(userID)
	return account, mapWalletTransportError(err)
}

func (a walletTransportAdapter) ListTransactions(userID uint, page, pageSize int) ([]models.WalletTransaction, int64, error) {
	transactions, total, err := a.wallets.ListTransactions(repository.WalletTransactionListFilter{
		Page: page, PageSize: pageSize, UserID: userID,
	})
	return transactions, total, mapWalletTransportError(err)
}

func (a walletTransportAdapter) ListAdminTransactions(userID uint, page, pageSize int, typ, direction string) ([]models.WalletTransaction, int64, error) {
	transactions, total, err := a.wallets.ListTransactions(repository.WalletTransactionListFilter{
		Page: page, PageSize: pageSize, UserID: userID, Type: typ, Direction: direction,
	})
	return transactions, total, mapWalletTransportError(err)
}

func (a walletTransportAdapter) ListRechargeOrdersAdmin(filter wallettransport.AdminRechargeListFilter) ([]models.WalletRechargeOrder, int64, error) {
	orders, total, err := a.wallets.ListRechargeOrdersAdmin(repository.WalletRechargeListFilter{
		Page:         filter.Page,
		PageSize:     filter.PageSize,
		RechargeNo:   filter.RechargeNo,
		UserID:       filter.UserID,
		UserKeyword:  filter.UserKeyword,
		PaymentID:    filter.PaymentID,
		ChannelID:    filter.ChannelID,
		ProviderType: filter.ProviderType,
		ChannelType:  filter.ChannelType,
		Status:       filter.Status,
		CreatedFrom:  filter.CreatedFrom,
		CreatedTo:    filter.CreatedTo,
		PaidFrom:     filter.PaidFrom,
		PaidTo:       filter.PaidTo,
	})
	return orders, total, mapWalletTransportError(err)
}

func (a walletTransportAdapter) AdminAdjustBalance(input wallettransport.AdjustBalanceInput) (*models.WalletAccount, *models.WalletTransaction, error) {
	account, txn, err := a.wallets.AdminAdjustBalance(service.WalletAdjustInput{
		UserID:   input.UserID,
		Delta:    input.Delta,
		Currency: input.Currency,
		Remark:   input.Remark,
	})
	return account, txn, mapWalletTransportError(err)
}

func (a walletTransportAdapter) ListUserRechargeOrders(userID uint, page, pageSize int, status, rechargeNo string) ([]models.WalletRechargeOrder, int64, error) {
	orders, total, err := a.wallets.ListUserRechargeOrders(userID, page, pageSize, status, rechargeNo)
	return orders, total, mapWalletTransportError(err)
}

func (a walletTransportAdapter) StatsUserRechargeOrders(userID uint, rechargeNo string) (map[string]int64, error) {
	stats, err := a.wallets.StatsUserRechargeOrders(userID, rechargeNo)
	return stats, mapWalletTransportError(err)
}

func (a walletTransportAdapter) GetRechargeOrderByRechargeNo(userID uint, rechargeNo string) (*models.WalletRechargeOrder, error) {
	order, err := a.wallets.GetRechargeOrderByRechargeNo(userID, rechargeNo)
	return order, mapWalletTransportError(err)
}

func (a walletTransportAdapter) GetRechargeOrderByPaymentIDAndUser(paymentID uint, userID uint) (*models.WalletRechargeOrder, error) {
	order, err := a.wallets.GetRechargeOrderByPaymentIDAndUser(paymentID, userID)
	return order, mapWalletTransportError(err)
}

func (a walletTransportAdapter) GetAvailableWalletRechargeChannels(amount money.Amount, user *userdomain.User) ([]map[string]interface{}, error) {
	channels, err := a.payments.GetAvailableChannels(service.AvailablePaymentChannelFilter{
		TargetAmount: &amount, User: user, PaymentType: constants.PaymentTypeWallet,
	})
	return channels, mapWalletTransportError(err)
}

func (a walletTransportAdapter) CreateWalletRechargePayment(input wallettransport.CreateRechargePaymentInput) (*wallettransport.CreateRechargePaymentResult, error) {
	result, err := a.payments.CreateWalletRechargePayment(service.CreateWalletRechargePaymentInput{
		UserID: input.UserID, ChannelID: input.ChannelID, Amount: input.Amount, Currency: input.Currency,
		Remark: input.Remark, ClientIP: input.ClientIP, Context: input.Context, RequestScheme: input.RequestScheme,
	})
	if err != nil {
		return nil, mapWalletTransportError(err)
	}
	return &wallettransport.CreateRechargePaymentResult{Recharge: result.Recharge, Payment: result.Payment}, nil
}

func (a walletTransportAdapter) GetPayment(id uint) (*models.Payment, error) {
	payment, err := a.payments.GetPayment(id)
	return payment, mapWalletTransportError(err)
}

func (a walletTransportAdapter) CapturePayment(input wallettransport.CapturePaymentInput) (*models.Payment, error) {
	payment, err := a.payments.CapturePayment(service.CapturePaymentInput{PaymentID: input.PaymentID, Context: input.Context})
	return payment, mapWalletTransportError(err)
}

func mapWalletTransportError(err error) error {
	if err == nil {
		return nil
	}
	for _, mapping := range []struct {
		source error
		target error
	}{
		{service.ErrWalletInvalidAmount, wallettransport.ErrInvalidAmount},
		{service.ErrWalletInsufficientBalance, wallettransport.ErrInsufficientBalance},
		{service.ErrWalletNotSupportedForGuest, wallettransport.ErrNotSupportedForGuest},
		{service.ErrWalletRechargeNotFound, wallettransport.ErrRechargeNotFound},
		{service.ErrPaymentInvalid, wallettransport.ErrPaymentInvalid},
		{service.ErrPaymentNotFound, wallettransport.ErrPaymentNotFound},
		{service.ErrOrderNotFound, wallettransport.ErrOrderNotFound},
		{service.ErrOrderStatusInvalid, wallettransport.ErrOrderStatusInvalid},
		{service.ErrPaymentChannelNotFound, wallettransport.ErrPaymentChannelNotFound},
		{service.ErrPaymentChannelInactive, wallettransport.ErrPaymentChannelInactive},
		{service.ErrPaymentProviderNotSupported, wallettransport.ErrPaymentProviderNotSupported},
		{service.ErrPaymentChannelConfigInvalid, wallettransport.ErrPaymentChannelConfigInvalid},
		{service.ErrPaymentGatewayRequestFailed, wallettransport.ErrPaymentGatewayRequestFailed},
		{service.ErrPaymentGatewayResponseInvalid, wallettransport.ErrPaymentGatewayResponseInvalid},
		{service.ErrPaymentCurrencyMismatch, wallettransport.ErrPaymentCurrencyMismatch},
		{service.ErrPaymentChannelNotAllowedForProduct, wallettransport.ErrPaymentChannelNotAllowedProduct},
		{service.ErrPaymentChannelNotAllowedForRecharge, wallettransport.ErrPaymentChannelNotAllowedRecharge},
		{service.ErrWalletOnlyPaymentRequired, wallettransport.ErrWalletOnlyPaymentRequired},
		{service.ErrPaymentStatusInvalid, wallettransport.ErrPaymentStatusInvalid},
		{service.ErrPaymentAmountMismatch, wallettransport.ErrPaymentAmountMismatch},
	} {
		if errors.Is(err, mapping.source) {
			return fmt.Errorf("%w: %v", mapping.target, err)
		}
	}
	return err
}
