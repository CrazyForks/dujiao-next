package reseller

import (
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/shopspring/decimal"
)

// OrderQueryService 分销销售订单只读查询用例。
type OrderQueryService struct {
	store OrderQueryStore
}

func NewOrderQueryService(store OrderQueryStore) *OrderQueryService {
	return &OrderQueryService{store: store}
}

type pricingItemSnapshot struct {
	BaseUnitAmount      string
	ResellerUnitAmount  string
	BaseTotalAmount     string
	ResellerTotalAmount string
	ProfitAmount        string
}

func (s *OrderQueryService) requireActiveProfileByUser(userID uint) (*models.ResellerProfile, error) {
	if s == nil || s.store == nil || userID == 0 {
		return nil, ErrNotOpened
	}
	profile, err := s.store.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrNotOpened
	}
	// 订单只读视角仅要求资料激活，不校验结算冻结（与提现/账务不同）。
	if profile.Status != models.ResellerProfileStatusActive {
		return nil, ErrProfileInactive
	}
	return profile, nil
}

func (s *OrderQueryService) ListUserOrders(userID uint, input OrderListInput) ([]OrderListItem, int64, error) {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return nil, 0, err
	}
	rows, total, err := s.store.ListOrderSnapshotsByReseller(orderSnapshotFilter(profile.ID, input))
	if err != nil {
		return nil, 0, err
	}
	out := make([]OrderListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, buildOrderListItem(row))
	}
	return out, total, nil
}

func (s *OrderQueryService) ListAdminOrders(resellerID uint, input OrderListInput) ([]OrderListItem, int64, error) {
	if s == nil || s.store == nil || resellerID == 0 {
		return nil, 0, catalogproduct.ErrNotFound
	}
	profile, err := s.store.GetProfileByID(resellerID)
	if err != nil {
		return nil, 0, err
	}
	if profile == nil {
		return nil, 0, catalogproduct.ErrNotFound
	}
	rows, total, err := s.store.ListOrderSnapshotsByReseller(orderSnapshotFilter(resellerID, input))
	if err != nil {
		return nil, 0, err
	}
	out := make([]OrderListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, buildOrderListItem(row))
	}
	return out, total, nil
}

func (s *OrderQueryService) GetUserOrderDetail(userID uint, orderNo string) (*OrderDetail, error) {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return nil, err
	}
	row, err := s.store.GetOrderSnapshotByResellerOrderNo(profile.ID, orderNo)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrOrderNotFound
	}
	detail := &OrderDetail{OrderListItem: buildOrderListItem(*row)}
	detail.Items = buildOrderItemDetails(*row)
	return detail, nil
}

func (s *OrderQueryService) StatsUserOrders(userID uint, input OrderListInput) (OrderStats, error) {
	profile, err := s.requireActiveProfileByUser(userID)
	if err != nil {
		return OrderStats{}, err
	}
	row, err := s.store.StatsOrderSnapshotsByReseller(orderSnapshotFilter(profile.ID, input))
	if err != nil {
		return OrderStats{}, err
	}
	return OrderStats(row), nil
}

func orderSnapshotFilter(resellerID uint, input OrderListInput) OrderSnapshotListFilter {
	return OrderSnapshotListFilter{
		ResellerID:  resellerID,
		Page:        input.Page,
		PageSize:    input.PageSize,
		Status:      strings.TrimSpace(input.Status),
		OrderNo:     strings.TrimSpace(input.OrderNo),
		CreatedFrom: input.CreatedFrom,
		CreatedTo:   input.CreatedTo,
		PaidFrom:    input.PaidFrom,
		PaidTo:      input.PaidTo,
	}
}

func buildOrderListItem(row OrderSnapshotRow) OrderListItem {
	order := row.Order
	snapshot := row.Snapshot
	return OrderListItem{
		OrderNo:      order.OrderNo,
		Status:       order.Status,
		Currency:     snapshot.Currency,
		TotalAmount:  order.TotalAmount,
		BaseAmount:   snapshot.BaseAmount,
		ProfitAmount: snapshot.ProfitAmount,
		ProfitStatus: neutralProfitStatus(snapshot, order, row.LedgerEntries),
		Domain:       snapshot.Domain,
		BuyerLabel:   maskBuyerLabel(order, row.BuyerEmail),
		ItemsCount:   len(row.Items),
		CreatedAt:    order.CreatedAt,
		PaidAt:       order.PaidAt,
	}
}

func neutralProfitStatus(snapshot models.ResellerOrderSnapshot, order models.Order, ledgerEntries []models.ResellerLedgerEntry) string {
	if !snapshot.ProfitEligible || snapshot.ProfitAmount.Decimal.LessThanOrEqual(decimal.Zero) {
		return ProfitStatusUnavailable
	}
	switch order.Status {
	case constants.OrderStatusCanceled, constants.OrderStatusRefunded, constants.OrderStatusPartiallyRefunded:
		return ProfitStatusUnavailable
	}
	if order.PaidAt == nil || order.Status == constants.OrderStatusPendingPayment {
		return ProfitStatusPending
	}
	for _, entry := range ledgerEntries {
		if entry.Type != models.ResellerLedgerTypeOrderProfit {
			continue
		}
		switch entry.Status {
		case models.ResellerLedgerStatusAvailable, models.ResellerLedgerStatusLocked, models.ResellerLedgerStatusWithdrawn:
			return ProfitStatusCredited
		case models.ResellerLedgerStatusPendingConfirm:
			return ProfitStatusPending
		case models.ResellerLedgerStatusCanceled:
			return ProfitStatusUnavailable
		}
	}
	return ProfitStatusPending
}

func maskBuyerLabel(order models.Order, buyerEmail string) string {
	if order.UserID > 0 {
		if label := maskBuyerEmail(buyerEmail); label != "" {
			return label
		}
		return fmt.Sprintf("user#%d", order.UserID)
	}
	if label := maskBuyerEmail(order.GuestEmail); label != "" {
		return label
	}
	return "guest"
}

func maskBuyerEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return ""
	}
	prefix := parts[0]
	if len(prefix) > 1 {
		prefix = prefix[:1]
	}
	return prefix + "***@" + parts[1]
}

func buildOrderItemDetails(row OrderSnapshotRow) []OrderItemDetail {
	pricingByItemID := pricingSnapshotByOrderItemID(row.Snapshot.PricingSnapshotJSON)
	out := make([]OrderItemDetail, 0, len(row.Items))
	for i := range row.Items {
		item := row.Items[i]
		itemPricing := pricingByItemID[item.ID]
		out = append(out, OrderItemDetail{
			Title:               item.TitleJSON,
			SKUSnapshot:         item.SKUSnapshotJSON,
			Quantity:            item.Quantity,
			UnitPrice:           item.UnitPrice,
			TotalPrice:          item.TotalPrice,
			BaseUnitAmount:      itemPricing.BaseUnitAmount,
			ResellerUnitAmount:  itemPricing.ResellerUnitAmount,
			BaseTotalAmount:     itemPricing.BaseTotalAmount,
			ResellerTotalAmount: itemPricing.ResellerTotalAmount,
			ProfitAmount:        itemPricing.ProfitAmount,
		})
	}
	return out
}

func pricingSnapshotByOrderItemID(snapshot models.JSON) map[uint]pricingItemSnapshot {
	out := map[uint]pricingItemSnapshot{}
	rawItems, ok := snapshot["items"].([]interface{})
	if !ok {
		return out
	}
	for _, raw := range rawItems {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		itemID := uintFromOrderSnapshotValue(item["order_item_id"])
		if itemID == 0 {
			continue
		}
		out[itemID] = pricingItemSnapshot{
			BaseUnitAmount:      orderSnapshotStringValue(item["base_unit_amount"]),
			ResellerUnitAmount:  orderSnapshotStringValue(item["reseller_unit_amount"]),
			BaseTotalAmount:     orderSnapshotStringValue(item["base_total_amount"]),
			ResellerTotalAmount: orderSnapshotStringValue(item["reseller_total_amount"]),
			ProfitAmount:        orderSnapshotStringValue(item["profit_amount"]),
		}
	}
	return out
}

func uintFromOrderSnapshotValue(value interface{}) uint {
	switch v := value.(type) {
	case uint:
		return v
	case int:
		if v > 0 {
			return uint(v)
		}
	case int64:
		if v > 0 {
			return uint(v)
		}
	case float64:
		if v > 0 {
			return uint(v)
		}
	case string:
		parsed, err := decimal.NewFromString(strings.TrimSpace(v))
		if err == nil && parsed.GreaterThan(decimal.Zero) {
			return uint(parsed.IntPart())
		}
	}
	return 0
}

func orderSnapshotStringValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case float64:
		return decimal.NewFromFloat(v).Round(2).StringFixed(2)
	case int:
		return decimal.NewFromInt(int64(v)).Round(2).StringFixed(2)
	case int64:
		return decimal.NewFromInt(v).Round(2).StringFixed(2)
	case uint:
		return decimal.NewFromInt(int64(v)).Round(2).StringFixed(2)
	}
	return ""
}
