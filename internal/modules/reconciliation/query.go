package reconciliation

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

func (s *Service) GetJob(id uint) (*models.ReconciliationJob, error) {
	return s.jobs.GetByID(id)
}

func (s *Service) ListJobs(filter JobListFilter) ([]models.ReconciliationJob, int64, error) {
	return s.jobs.List(filter)
}

func (s *Service) GetJobItems(jobID uint, page, pageSize int) ([]models.ReconciliationItem, int64, error) {
	return s.items.ListByJobID(jobID, page, pageSize)
}

func (s *Service) ResolveItem(itemID, adminID uint, remark string) error {
	item, err := s.items.GetByID(itemID)
	if err != nil || item == nil {
		return ErrItemNotFound
	}
	now := time.Now()
	item.Resolved, item.ResolvedBy, item.ResolvedAt, item.Remark = true, &adminID, &now, remark
	return s.items.Update(item)
}
