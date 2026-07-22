package repository

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// CreateOrderSnapshot 创建订单分销快照。
func (r *GormResellerRepository) CreateOrderSnapshot(snapshot *models.ResellerOrderSnapshot) error {
	if snapshot == nil || snapshot.OrderID == 0 || snapshot.ResellerID == 0 {
		return errors.New("invalid reseller order snapshot")
	}
	profitEligible := snapshot.ProfitEligible
	if err := r.db.Create(snapshot).Error; err != nil {
		return err
	}
	if !profitEligible {
		if err := r.db.Model(&models.ResellerOrderSnapshot{}).
			Where("id = ?", snapshot.ID).
			Update("profit_eligible", false).Error; err != nil {
			return err
		}
		snapshot.ProfitEligible = false
	}
	return nil
}

// GetOrderSnapshotByOrderID 按订单 ID 获取订单分销快照。
func (r *GormResellerRepository) GetOrderSnapshotByOrderID(orderID uint) (*models.ResellerOrderSnapshot, error) {
	if orderID == 0 {
		return nil, nil
	}
	var snapshot models.ResellerOrderSnapshot
	if err := r.db.Where("order_id = ?", orderID).First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &snapshot, nil
}

func applyResellerOrderSnapshotFilter(query *gorm.DB, filter ResellerOrderListFilter) *gorm.DB {
	query = query.Joins("JOIN orders ON orders.id = reseller_order_snapshots.order_id AND orders.deleted_at IS NULL").
		Where("reseller_order_snapshots.deleted_at IS NULL").
		Where("reseller_order_snapshots.reseller_id = ?", filter.ResellerID).
		Where("orders.parent_id IS NULL").
		Where("orders.reseller_id = reseller_order_snapshots.reseller_id")
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("orders.status = ?", status)
	}
	if orderNo := strings.TrimSpace(filter.OrderNo); orderNo != "" {
		query = query.Where("orders.order_no LIKE ?", "%"+orderNo+"%")
	}
	if domain := strings.TrimSpace(filter.Domain); domain != "" {
		query = query.Where("reseller_order_snapshots.domain = ?", domain)
	}
	if currency := strings.TrimSpace(filter.Currency); currency != "" {
		query = query.Where("reseller_order_snapshots.currency = ?", currency)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("orders.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("orders.created_at <= ?", *filter.CreatedTo)
	}
	if filter.PaidFrom != nil {
		query = query.Where("orders.paid_at >= ?", *filter.PaidFrom)
	}
	if filter.PaidTo != nil {
		query = query.Where("orders.paid_at <= ?", *filter.PaidTo)
	}
	return query
}

// ListOrderSnapshotsByReseller 分页列出分销商订单快照并补齐订单展示数据。
func (r *GormResellerRepository) ListOrderSnapshotsByReseller(filter ResellerOrderListFilter) ([]ResellerOrderSnapshotRow, int64, error) {
	rows := make([]models.ResellerOrderSnapshot, 0)
	if filter.ResellerID == 0 {
		return []ResellerOrderSnapshotRow{}, 0, nil
	}
	query := applyResellerOrderSnapshotFilter(r.db.Model(&models.ResellerOrderSnapshot{}), filter)
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Select("reseller_order_snapshots.*").
		Preload("Order.Items").
		Preload("Order.Children.Items").
		Order("orders.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out, err := r.buildResellerOrderSnapshotRows(filter.ResellerID, rows)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// StatsOrderSnapshotsByReseller 汇总分销商订单快照统计。
func (r *GormResellerRepository) StatsOrderSnapshotsByReseller(filter ResellerOrderListFilter) (ResellerOrderStatsRow, error) {
	result := ResellerOrderStatsRow{ByStatus: map[string]int64{}, ByCurrency: map[string]int64{}}
	if filter.ResellerID == 0 {
		return result, nil
	}
	query := applyResellerOrderSnapshotFilter(r.db.Model(&models.ResellerOrderSnapshot{}), filter)
	type statusRow struct {
		Status string
		Count  int64
	}
	var statuses []statusRow
	if err := query.Session(&gorm.Session{}).
		Select("orders.status AS status, COUNT(1) AS count").
		Group("orders.status").
		Scan(&statuses).Error; err != nil {
		return result, err
	}
	for _, row := range statuses {
		result.ByStatus[row.Status] = row.Count
		result.Total += row.Count
	}
	type currencyRow struct {
		Currency string
		Count    int64
	}
	var currencies []currencyRow
	if err := query.Session(&gorm.Session{}).
		Select("reseller_order_snapshots.currency AS currency, COUNT(1) AS count").
		Group("reseller_order_snapshots.currency").
		Scan(&currencies).Error; err != nil {
		return result, err
	}
	for _, row := range currencies {
		result.ByCurrency[row.Currency] = row.Count
	}
	return result, nil
}

// GetOrderSnapshotByResellerOrderNo 按订单号获取分销商自己的订单快照。
func (r *GormResellerRepository) GetOrderSnapshotByResellerOrderNo(resellerID uint, orderNo string) (*ResellerOrderSnapshotRow, error) {
	orderNo = strings.TrimSpace(orderNo)
	if resellerID == 0 || orderNo == "" {
		return nil, nil
	}
	var snapshot models.ResellerOrderSnapshot
	err := r.db.Model(&models.ResellerOrderSnapshot{}).
		Joins("JOIN orders ON orders.id = reseller_order_snapshots.order_id AND orders.deleted_at IS NULL").
		Select("reseller_order_snapshots.*").
		Where("reseller_order_snapshots.deleted_at IS NULL").
		Where("reseller_order_snapshots.reseller_id = ?", resellerID).
		Where("orders.parent_id IS NULL").
		Where("orders.reseller_id = reseller_order_snapshots.reseller_id").
		Where("orders.order_no = ?", orderNo).
		Preload("Order.Items").
		Preload("Order.Children.Items").
		First(&snapshot).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	rows, err := r.buildResellerOrderSnapshotRows(resellerID, []models.ResellerOrderSnapshot{snapshot})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (r *GormResellerRepository) buildResellerOrderSnapshotRows(resellerID uint, snapshots []models.ResellerOrderSnapshot) ([]ResellerOrderSnapshotRow, error) {
	orderIDs := make([]uint, 0, len(snapshots))
	buyerUserIDs := make([]uint, 0, len(snapshots))
	buyerSeen := map[uint]struct{}{}
	for i := range snapshots {
		if snapshots[i].OrderID > 0 {
			orderIDs = append(orderIDs, snapshots[i].OrderID)
		}
		buyerUserID := snapshots[i].Order.UserID
		if buyerUserID > 0 {
			if _, ok := buyerSeen[buyerUserID]; !ok {
				buyerSeen[buyerUserID] = struct{}{}
				buyerUserIDs = append(buyerUserIDs, buyerUserID)
			}
		}
	}
	ledgerByOrderID := map[uint][]models.ResellerLedgerEntry{}
	if len(orderIDs) > 0 {
		var ledgerRows []models.ResellerLedgerEntry
		if err := r.db.Where("reseller_id = ? AND order_id IN ?", resellerID, orderIDs).
			Order("id DESC").
			Find(&ledgerRows).Error; err != nil {
			return nil, err
		}
		for i := range ledgerRows {
			if ledgerRows[i].OrderID == nil {
				continue
			}
			ledgerByOrderID[*ledgerRows[i].OrderID] = append(ledgerByOrderID[*ledgerRows[i].OrderID], ledgerRows[i])
		}
	}
	buyerEmailByID := map[uint]string{}
	if len(buyerUserIDs) > 0 {
		var users []models.User
		if err := r.db.Select("id", "email").Where("id IN ?", buyerUserIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for i := range users {
			buyerEmailByID[users[i].ID] = users[i].Email
		}
	}
	out := make([]ResellerOrderSnapshotRow, 0, len(snapshots))
	for i := range snapshots {
		items := resellerOrderItemsFromParentOrChildren(snapshots[i].Order)
		out = append(out, ResellerOrderSnapshotRow{
			Snapshot:      snapshots[i],
			Order:         snapshots[i].Order,
			Items:         items,
			LedgerEntries: ledgerByOrderID[snapshots[i].OrderID],
			BuyerEmail:    buyerEmailByID[snapshots[i].Order.UserID],
		})
	}
	return out, nil
}

func resellerOrderItemsFromParentOrChildren(order models.Order) []models.OrderItem {
	if len(order.Items) > 0 {
		return order.Items
	}
	items := make([]models.OrderItem, 0)
	for i := range order.Children {
		items = append(items, order.Children[i].Items...)
	}
	return items
}
