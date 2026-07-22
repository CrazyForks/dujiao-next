package fulfillmentwiring

import (
	"errors"
	"fmt"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	fulfillmenttransport "github.com/dujiao-next/internal/transport/http/fulfillment"
)

type fulfillmentManualCreatorAdapter struct {
	svc *service.FulfillmentService
}

func (a fulfillmentManualCreatorAdapter) CreateManual(input fulfillmenttransport.CreateManualInput) (*models.Fulfillment, error) {
	res, err := a.svc.CreateManual(service.CreateManualInput{
		OrderID:      input.OrderID,
		AdminID:      input.AdminID,
		Payload:      input.Payload,
		DeliveryData: input.DeliveryData,
	})
	return res, mapFulfillmentTransportError(err)
}

type fulfillmentAdminOrderAdapter struct {
	orders *service.OrderService
}

func (a fulfillmentAdminOrderAdapter) GetOrderForAdmin(orderID uint) (*models.Order, error) {
	order, err := a.orders.GetOrderForAdmin(orderID)
	return order, mapFulfillmentTransportError(err)
}

func mapFulfillmentTransportError(err error) error {
	if err == nil {
		return nil
	}
	for _, mapping := range []struct {
		source error
		target error
	}{
		{service.ErrFulfillmentExists, fulfillmenttransport.ErrFulfillmentExists},
		{service.ErrFulfillmentInvalid, fulfillmenttransport.ErrFulfillmentInvalid},
		{service.ErrOrderStatusInvalid, fulfillmenttransport.ErrOrderStatusInvalid},
		{service.ErrOrderNotFound, fulfillmenttransport.ErrOrderNotFound},
	} {
		if errors.Is(err, mapping.source) {
			return fmt.Errorf("%w: %v", mapping.target, err)
		}
	}
	return err
}
