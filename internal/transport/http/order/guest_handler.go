package orderhttp

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/reseller"

	"github.com/gin-gonic/gin"
)

// GuestOrderQuery 游客订单只读端口。
type GuestOrderQuery interface {
	ListOrdersByGuestForTenant(tenant reseller.TenantContext, email, password string, page, pageSize int) ([]models.Order, int64, error)
	GetOrderByGuestOrderNoForTenant(tenant reseller.TenantContext, orderNo, email, password string) (*models.Order, error)
	GetAnyOrderByGuestOrderNoForTenant(tenant reseller.TenantContext, orderNo, email, password string) (*models.Order, error)
}

// GuestHandler 处理前台游客订单只读 HTTP。
type GuestHandler struct {
	orders   GuestOrderQuery
	payments PaymentChannelPolicy
	refunds  RefundRecordDirectory
}

func NewGuestHandler(orders GuestOrderQuery, payments PaymentChannelPolicy, refunds RefundRecordDirectory) *GuestHandler {
	if orders == nil {
		panic("order guest handler: orders is nil")
	}
	return &GuestHandler{orders: orders, payments: payments, refunds: refunds}
}

// ListGuestOrders 获取游客订单列表
func (h *GuestHandler) ListGuestOrders(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	password := strings.TrimSpace(c.Query("order_password"))
	orderNo := strings.TrimSpace(c.Query("order_no"))
	if email == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	if password == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.guest_password_required", nil)
		return
	}

	if orderNo != "" {
		order, err := h.orders.GetOrderByGuestOrderNoForTenant(tenantFromRequest(c), orderNo, email, password)
		if err != nil {
			if errors.Is(err, ErrGuestOrderNotFound) {
				pagination := response.Pagination{
					Page:      1,
					PageSize:  1,
					Total:     0,
					TotalPage: 1,
				}
				response.SuccessWithPage(c, []models.Order{}, pagination)
				return
			}
			shared.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
			return
		}
		pagination := response.Pagination{
			Page:      1,
			PageSize:  1,
			Total:     1,
			TotalPage: 1,
		}
		response.SuccessWithPage(c, dto.NewOrderSummaryList([]models.Order{*order}), pagination)
		return
	}

	page, pageSize := shared.ParsePagination(c)

	orders, total, err := h.orders.ListOrdersByGuestForTenant(tenantFromRequest(c), email, password, page, pageSize)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	pagination := response.BuildPagination(page, pageSize, total)
	response.SuccessWithPage(c, dto.NewOrderSummaryList(orders), pagination)
}

// GetGuestOrderByOrderNo 按订单号获取游客订单详情
func (h *GuestHandler) GetGuestOrderByOrderNo(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	password := strings.TrimSpace(c.Query("order_password"))
	if email == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	if password == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.guest_password_required", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.order_item_invalid", nil)
		return
	}
	order, err := h.orders.GetOrderByGuestOrderNoForTenant(tenantFromRequest(c), orderNo, email, password)
	if err != nil {
		if errors.Is(err, ErrGuestOrderNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.guest_order_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	orderDetail := dto.NewOrderDetailTruncated(order)
	enrichOrderWithAllowedChannels(h.payments, order, &orderDetail)
	enrichOrderWithRefundRecords(h.refunds, order, &orderDetail)
	response.Success(c, orderDetail)
}

// DownloadGuestFulfillment 下载订单交付内容（游客）
// 支持父订单或子订单的 order_no
func (h *GuestHandler) DownloadGuestFulfillment(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	password := strings.TrimSpace(c.Query("order_password"))
	if email == "" || password == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.order_item_invalid", nil)
		return
	}
	order, err := h.orders.GetAnyOrderByGuestOrderNoForTenant(tenantFromRequest(c), orderNo, email, password)
	if err != nil {
		if errors.Is(err, ErrGuestOrderNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.guest_order_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	if order == nil {
		shared.RespondError(c, response.CodeNotFound, "error.guest_order_not_found", nil)
		return
	}
	respondFulfillmentDownload(c, order)
}
