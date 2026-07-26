package paymentwiring

import (
	"github.com/dujiao-next/internal/provider"
	paymenttransport "github.com/dujiao-next/internal/transport/http/payment"
	paymentcallbacktransport "github.com/dujiao-next/internal/transport/http/payment/callback"
)

// Handlers is the complete payment HTTP entrypoint set assembled at the
// application boundary.
type Handlers struct {
	Latest       *paymenttransport.LatestHandler
	Write        *paymenttransport.WriteHandler
	Admin        *paymenttransport.AdminHandler
	AdminChannel *paymenttransport.AdminChannelHandler
	Webhook      *paymenttransport.WebhookHandler
	Callback     *paymentcallbacktransport.Handler
}

// New assembles payment transports without exposing legacy adapters to router.
func New(c *provider.Container) Handlers {
	guestOrders := guestOrderLookupAdapter{orders: c.OrderService}
	userOrders := userOrderLookupAdapter{orders: c.OrderService}
	alerter := exceptionAlerterAdapter{notifications: c.NotificationService}

	return Handlers{
		Latest: paymenttransport.NewLatestHandler(
			guestOrders,
			userOrders,
			pendingLookupAdapter{payments: c.PaymentRepo},
		),
		Write: paymenttransport.NewWriteHandler(
			guestOrders,
			userOrders,
			writerAdapter{payments: c.PaymentService},
		),
		Admin: paymenttransport.NewAdminHandler(
			adminQueryAdapter{payments: c.PaymentService},
			adminChannelLookupAdapter{channels: c.PaymentChannelRepo},
			adminOrderLookupAdapter{orders: c.OrderStore},
			adminRechargeLookupAdapter{wallets: c.WalletRepo},
		),
		AdminChannel: paymenttransport.NewAdminChannelHandler(
			adminChannelCatalogAdapter{payments: c.PaymentService, channels: c.PaymentChannelRepo},
		),
		Webhook: paymenttransport.NewWebhookHandler(
			webhookServiceAdapter{payments: c.PaymentService},
			alerter,
		),
		Callback: paymentcallbacktransport.NewHandler(
			callbackServiceAdapter{payments: c.PaymentService},
			c.PaymentRepo,
			c.PaymentChannelRepo,
			alerter,
		),
	}
}
