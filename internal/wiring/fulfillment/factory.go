package fulfillmentwiring

import (
	"github.com/dujiao-next/internal/provider"
	fulfillmenttransport "github.com/dujiao-next/internal/transport/http/fulfillment"
)

func NewAdminHandler(c *provider.Container) *fulfillmenttransport.AdminHandler {
	return fulfillmenttransport.NewAdminHandler(
		fulfillmentManualCreatorAdapter{svc: c.FulfillmentService},
		fulfillmentAdminOrderAdapter{orders: c.OrderService},
	)
}
