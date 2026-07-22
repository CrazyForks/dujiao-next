package reconciliation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/notification"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/upstream"

	"github.com/hibiken/asynq"
)

var (
	ErrJobNotFound  = errors.New("reconciliation job not found")
	ErrItemNotFound = errors.New("reconciliation item not found")
	ErrJobRunning   = errors.New("reconciliation job is already running")
)

type JobListFilter struct {
	Page         int
	PageSize     int
	ConnectionID uint
	Status       string
	Type         string
}

type RunInput struct {
	ConnectionID   uint      `json:"connection_id" binding:"required"`
	Type           string    `json:"type" binding:"required"`
	TimeRangeStart time.Time `json:"time_range_start" binding:"required"`
	TimeRangeEnd   time.Time `json:"time_range_end" binding:"required"`
}

type JobRepository interface {
	Create(job *models.ReconciliationJob) error
	GetByID(id uint) (*models.ReconciliationJob, error)
	Update(job *models.ReconciliationJob) error
	List(filter JobListFilter) ([]models.ReconciliationJob, int64, error)
}

type ItemRepository interface {
	BatchCreate(items []models.ReconciliationItem) error
	GetByID(id uint) (*models.ReconciliationItem, error)
	Update(item *models.ReconciliationItem) error
	ListByJobID(jobID uint, page, pageSize int) ([]models.ReconciliationItem, int64, error)
}

type ProcurementReader interface {
	ListByConnectionAndTimeRange(connectionID uint, start, end time.Time) ([]models.ProcurementOrder, error)
}

type ConnectionProvider interface {
	GetByID(id uint) (*models.SiteConnection, error)
	GetAdapter(connection *models.SiteConnection) (upstream.Adapter, error)
}

type Enqueuer interface {
	EnqueueReconciliationRun(payload queue.ReconciliationRunPayload, opts ...asynq.Option) error
}

type NotificationEnqueuer interface {
	Enqueue(input notification.EnqueueInput) error
}

type ServiceOptions struct {
	Jobs          JobRepository
	Items         ItemRepository
	Procurements  ProcurementReader
	Connections   ConnectionProvider
	Queue         Enqueuer
	Notifications NotificationEnqueuer
}

type Service struct {
	jobs          JobRepository
	items         ItemRepository
	procurements  ProcurementReader
	connections   ConnectionProvider
	queue         Enqueuer
	notifications NotificationEnqueuer
}

func NewService(options ServiceOptions) *Service {
	if options.Jobs == nil || options.Items == nil || options.Procurements == nil || options.Connections == nil {
		panic("reconciliation service: required dependency is nil")
	}
	return &Service{
		jobs: options.Jobs, items: options.Items, procurements: options.Procurements,
		connections: options.Connections, queue: options.Queue, notifications: options.Notifications,
	}
}

func (s *Service) CreateAndEnqueue(input RunInput) (*models.ReconciliationJob, error) {
	job := &models.ReconciliationJob{
		ConnectionID: input.ConnectionID, Type: input.Type,
		Status:         constants.ReconciliationJobStatusPending,
		TimeRangeStart: input.TimeRangeStart, TimeRangeEnd: input.TimeRangeEnd,
	}
	if err := s.jobs.Create(job); err != nil {
		return nil, fmt.Errorf("create reconciliation job: %w", err)
	}
	if s.queue != nil {
		if err := s.queue.EnqueueReconciliationRun(queue.ReconciliationRunPayload{JobID: job.ID}); err != nil {
			logger.Warnw("reconciliation_enqueue_failed", "job_id", job.ID, "error", err)
		}
	}
	return job, nil
}

func (s *Service) Execute(ctx context.Context, jobID uint) error {
	job, err := s.jobs.GetByID(jobID)
	if err != nil {
		return fmt.Errorf("get reconciliation job: %w", err)
	}
	if job == nil {
		return ErrJobNotFound
	}
	if job.Status == constants.ReconciliationJobStatusRunning {
		return ErrJobRunning
	}
	if job.Status == constants.ReconciliationJobStatusCompleted {
		return nil
	}

	now := time.Now()
	job.Status, job.StartedAt = constants.ReconciliationJobStatusRunning, &now
	if err := s.jobs.Update(job); err != nil {
		return fmt.Errorf("update job status to running: %w", err)
	}
	if err := s.execute(ctx, job); err != nil {
		finishedAt := time.Now()
		job.Status, job.FinishedAt = constants.ReconciliationJobStatusFailed, &finishedAt
		job.ResultJSON = marshalResult(map[string]string{"error": err.Error()})
		_ = s.jobs.Update(job)
		return fmt.Errorf("execute reconciliation: %w", err)
	}

	finishedAt := time.Now()
	job.Status, job.FinishedAt = constants.ReconciliationJobStatusCompleted, &finishedAt
	_ = s.jobs.Update(job)
	if job.MismatchedCount > 0 {
		s.notifyMismatch(job)
	}
	return nil
}
