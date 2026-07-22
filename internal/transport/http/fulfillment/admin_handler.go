package fulfillmenthttp

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

var (
	ErrFulfillmentExists  = errors.New("fulfillment exists")
	ErrFulfillmentInvalid = errors.New("fulfillment invalid")
	ErrOrderStatusInvalid = errors.New("order status invalid")
	ErrOrderNotFound      = errors.New("order not found")
)

// CreateManualInput 创建人工交付输入。
type CreateManualInput struct {
	OrderID      uint
	AdminID      uint
	Payload      string
	DeliveryData models.JSON
}

// ManualCreator 管理端录入交付端口。
type ManualCreator interface {
	CreateManual(input CreateManualInput) (*models.Fulfillment, error)
}

// AdminOrderReader 管理端订单读取端口（下载交付用）。
type AdminOrderReader interface {
	GetOrderForAdmin(orderID uint) (*models.Order, error)
}

// AdminHandler 处理后台交付相关 HTTP 请求。
type AdminHandler struct {
	creator ManualCreator
	orders  AdminOrderReader
}

func NewAdminHandler(creator ManualCreator, orders AdminOrderReader) *AdminHandler {
	if creator == nil {
		panic("fulfillment admin handler: creator is nil")
	}
	if orders == nil {
		panic("fulfillment admin handler: orders is nil")
	}
	return &AdminHandler{creator: creator, orders: orders}
}

// AdminCreateFulfillmentRequest 管理端录入交付请求。
type AdminCreateFulfillmentRequest struct {
	OrderID      uint        `json:"order_id" binding:"required"`
	Payload      string      `json:"payload"`
	DeliveryData models.JSON `json:"delivery_data"`
}

// AdminCreateFulfillment 管理端录入交付内容。
func (h *AdminHandler) AdminCreateFulfillment(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	var req AdminCreateFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	fulfillment, err := h.creator.CreateManual(CreateManualInput{
		OrderID:      req.OrderID,
		AdminID:      adminID,
		Payload:      req.Payload,
		DeliveryData: req.DeliveryData,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrFulfillmentExists):
			shared.RespondError(c, response.CodeBadRequest, "error.fulfillment_exists", nil)
		case errors.Is(err, ErrFulfillmentInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.fulfillment_invalid", nil)
		case errors.Is(err, ErrOrderStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.order_status_invalid", nil)
		case errors.Is(err, ErrOrderNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.order_not_found", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.fulfillment_create_failed", err)
		}
		return
	}

	response.Success(c, fulfillment)
}

// AdminDownloadFulfillment 管理端下载订单交付内容。
func (h *AdminHandler) AdminDownloadFulfillment(c *gin.Context) {
	orderID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.order_item_invalid", nil)
		return
	}
	order, err := h.orders.GetOrderForAdmin(orderID)
	if err != nil || order == nil {
		shared.RespondError(c, response.CodeNotFound, "error.order_not_found", nil)
		return
	}
	payload := collectAdminFulfillmentPayload(order)
	if payload == "" {
		shared.RespondError(c, response.CodeNotFound, "error.fulfillment_not_found", nil)
		return
	}
	filename := "fulfillment-" + order.OrderNo + ".txt"
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(200, "text/plain; charset=utf-8", []byte(payload))
}

func collectAdminFulfillmentPayload(order *models.Order) string {
	if order.Fulfillment != nil && order.Fulfillment.Payload != "" {
		return order.Fulfillment.Payload
	}
	var parts []string
	for _, child := range order.Children {
		if child.Fulfillment != nil && child.Fulfillment.Payload != "" {
			parts = append(parts, child.Fulfillment.Payload)
		}
	}
	return strings.Join(parts, "\n")
}
