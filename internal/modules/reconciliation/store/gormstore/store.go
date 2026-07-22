package gormstore

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/reconciliation"

	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(job *models.ReconciliationJob) error {
	return s.db.Create(job).Error
}

func (s *Store) GetByID(id uint) (*models.ReconciliationJob, error) {
	var job models.ReconciliationJob
	if err := s.db.Preload("Connection").First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *Store) Update(job *models.ReconciliationJob) error {
	return s.db.Save(job).Error
}

func (s *Store) List(filter reconciliation.JobListFilter) ([]models.ReconciliationJob, int64, error) {
	var jobs []models.ReconciliationJob
	var total int64
	query := s.db.Model(&models.ReconciliationJob{})
	if filter.ConnectionID > 0 {
		query = query.Where("connection_id = ?", filter.ConnectionID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := filter.Page, filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if err := query.Preload("Connection").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (s *Store) BatchCreate(items []models.ReconciliationItem) error {
	if len(items) == 0 {
		return nil
	}
	return s.db.Create(&items).Error
}

func (s *Store) GetItemByID(id uint) (*models.ReconciliationItem, error) {
	var item models.ReconciliationItem
	if err := s.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) UpdateItem(item *models.ReconciliationItem) error {
	return s.db.Save(item).Error
}

func (s *Store) ListByJobID(jobID uint, page, pageSize int) ([]models.ReconciliationItem, int64, error) {
	var items []models.ReconciliationItem
	var total int64
	query := s.db.Model(&models.ReconciliationItem{}).Where("job_id = ?", jobID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ItemStore adapts the shared GORM store to the item repository's method names.
type ItemStore struct {
	store *Store
}

func NewItemStore(store *Store) *ItemStore {
	return &ItemStore{store: store}
}

func (s *ItemStore) BatchCreate(items []models.ReconciliationItem) error {
	return s.store.BatchCreate(items)
}

func (s *ItemStore) GetByID(id uint) (*models.ReconciliationItem, error) {
	return s.store.GetItemByID(id)
}

func (s *ItemStore) Update(item *models.ReconciliationItem) error {
	return s.store.UpdateItem(item)
}

func (s *ItemStore) ListByJobID(jobID uint, page, pageSize int) ([]models.ReconciliationItem, int64, error) {
	return s.store.ListByJobID(jobID, page, pageSize)
}
