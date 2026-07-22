package gormstore

import (
	"errors"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/procurement"

	"gorm.io/gorm"
)

// Store 是采购单持久化端口的 GORM 实现。
type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// GetByID 根据 ID 获取
func (r *Store) GetByID(id uint) (*models.ProcurementOrder, error) {
	var order models.ProcurementOrder
	if err := r.db.Preload("Connection", "deleted_at IS NULL").Preload("LocalOrder").Preload("LocalOrder.Items").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByLocalOrderID 根据本地订单 ID 获取
func (r *Store) GetByLocalOrderID(localOrderID uint) (*models.ProcurementOrder, error) {
	var order models.ProcurementOrder
	if err := r.db.Where("local_order_id = ?", localOrderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByLocalOrderNo 根据本地订单号获取
func (r *Store) GetByLocalOrderNo(localOrderNo string) (*models.ProcurementOrder, error) {
	var order models.ProcurementOrder
	if err := r.db.Where("local_order_no = ?", localOrderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// Create 创建采购单
func (r *Store) Create(order *models.ProcurementOrder) error {
	return r.db.Create(order).Error
}

// UpdateStatus 更新采购单状态
func (r *Store) UpdateStatus(id uint, status string, updates map[string]interface{}) error {
	if updates == nil {
		updates = map[string]interface{}{}
	}
	updates["status"] = status
	return r.db.Model(&models.ProcurementOrder{}).Where("id = ?", id).Updates(updates).Error
}

// List 列表查询
func (r *Store) List(filter procurement.ListFilter) ([]models.ProcurementOrder, int64, error) {
	var orders []models.ProcurementOrder
	var total int64

	q := r.db.Model(&models.ProcurementOrder{})
	if filter.ConnectionID > 0 {
		q = q.Where("connection_id = ?", filter.ConnectionID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.LocalOrderNo != "" {
		q = q.Where("local_order_no = ?", filter.LocalOrderNo)
	}
	if filter.UpstreamOrderNo != "" {
		q = q.Where("upstream_order_no = ?", filter.UpstreamOrderNo)
	}
	if filter.CreatedFrom != nil {
		q = q.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		q = q.Where("created_at <= ?", *filter.CreatedTo)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = q.Order("created_at DESC").Preload("Connection", "deleted_at IS NULL").Preload("LocalOrder").Preload("LocalOrder.Items")
	if filter.Page > 0 && filter.PageSize > 0 {
		q = q.Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize)
	}

	if err := q.Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// StatsByStatus 按状态聚合采购单数量（忽略分页，复用其他筛选条件）
func (r *Store) StatsByStatus(filter procurement.ListFilter) (map[string]int64, error) {
	q := r.db.Model(&models.ProcurementOrder{})
	if filter.ConnectionID > 0 {
		q = q.Where("connection_id = ?", filter.ConnectionID)
	}
	// 注意：StatsByStatus 不应用 filter.Status，因为聚合的目的就是看各状态分布
	if filter.LocalOrderNo != "" {
		q = q.Where("local_order_no = ?", filter.LocalOrderNo)
	}
	if filter.UpstreamOrderNo != "" {
		q = q.Where("upstream_order_no = ?", filter.UpstreamOrderNo)
	}
	if filter.CreatedFrom != nil {
		q = q.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		q = q.Where("created_at <= ?", *filter.CreatedTo)
	}

	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := q.Select("status, COUNT(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(rows))
	for _, item := range rows {
		result[item.Status] = item.Count
	}
	return result, nil
}

// ListByConnectionAndTimeRange 按连接和时间范围查询采购单
func (r *Store) ListByConnectionAndTimeRange(connectionID uint, start, end time.Time) ([]models.ProcurementOrder, error) {
	var orders []models.ProcurementOrder
	q := r.db.Preload("LocalOrder").Where("connection_id = ? AND created_at >= ? AND created_at <= ?", connectionID, start, end)
	if err := q.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
