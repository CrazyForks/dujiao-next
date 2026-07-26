package fulfillmentwiring

import (
	fulfillmenttransport "github.com/dujiao-next/internal/modules/fulfillment/transport/http"
	"github.com/dujiao-next/internal/provider"
)

func NewAdminHandler(c *provider.Container) *fulfillmenttransport.AdminHandler {
	return fulfillmenttransport.NewAdminHandler(
		fulfillmentManualCreatorAdapter{svc: c.FulfillmentService},
		fulfillmentAdminOrderAdapter{orders: c.OrderService},
	)
}
