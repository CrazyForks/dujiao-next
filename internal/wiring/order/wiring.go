package orderwiring

import (
	"github.com/dujiao-next/internal/provider"
	ordertransport "github.com/dujiao-next/internal/transport/http/order"
)

// Handlers contains every order HTTP entrypoint required by the router.
type Handlers struct {
	Admin       *ordertransport.AdminHandler
	AdminRefund *ordertransport.AdminRefundHandler
	User        *ordertransport.UserHandler
	Guest       *ordertransport.GuestHandler
	Preview     *ordertransport.PreviewHandler
	Create      *ordertransport.CreateHandler
}

// New assembles order transports and their legacy boundary adapters.
func New(c *provider.Container) Handlers {
	return Handlers{
		Admin: ordertransport.NewAdminHandler(
			orderAdminQueryAdapter{orders: c.OrderService},
			orderAdminUserAdapter{users: c.UserRepo},
			orderAdminCouponAdapter{coupons: c.CouponRepo},
			orderAdminPromotionAdapter{promotions: c.PromotionRepo},
			orderAdminPaymentAdapter{payments: c.PaymentRepo},
			orderAdminPaymentChannelAdapter{channels: c.PaymentChannelRepo},
		),
		AdminRefund: NewAdminRefundHandler(c),
		User: ordertransport.NewUserHandler(
			orderUserQueryAdapter{orders: c.OrderService},
			orderUserPaymentChannelAdapter{payments: c.PaymentService},
			orderUserRefundRecordAdapter{records: c.OrderRefundRecordRepo},
			orderUserLookupAdapter{users: c.UserRepo},
		),
		Guest: ordertransport.NewGuestHandler(
			orderGuestQueryAdapter{orders: c.OrderService},
			orderUserPaymentChannelAdapter{payments: c.PaymentService},
			orderUserRefundRecordAdapter{records: c.OrderRefundRecordRepo},
		),
		Preview: ordertransport.NewPreviewHandler(
			orderPreviewAdapter{orders: c.OrderService},
		),
		Create: ordertransport.NewCreateHandler(
			orderCreateAdapter{orders: c.OrderService},
			orderUserPaymentChannelAdapter{payments: c.PaymentService},
			orderGuestCreateCaptchaAdapter{captcha: c.CaptchaService},
			orderPaymentCreatorAdapter{payments: c.PaymentService},
		),
	}
}

// NewAdminRefundHandler exposes the focused refund composition for integration
// tests and command surfaces that do not need the complete order handler set.
func NewAdminRefundHandler(c *provider.Container) *ordertransport.AdminRefundHandler {
	refunds := orderAdminRefundAdapter{refunds: c.OrderRefundService}
	return ordertransport.NewAdminRefundHandler(
		refunds,
		refunds,
		orderAdminWalletRefundAdapter{wallets: c.WalletService},
		orderAdminOrderLookupAdapter{orders: c.OrderRepo},
		orderAdminStatusEmailAdapter{queue: c.QueueClient},
	)
}
