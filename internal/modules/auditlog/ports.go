package auditlog

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

type UserLoginFilter struct {
	Page        int
	PageSize    int
	UserID      uint
	Email       string
	Status      string
	FailReason  string
	ClientIP    string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type AuthzFilter struct {
	Page            int
	PageSize        int
	OperatorAdminID uint
	TargetAdminID   uint
	Action          string
	Role            string
	Object          string
	Method          string
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
}

type AdminLoginFilter struct {
	Page      int
	PageSize  int
	AdminID   *uint
	Username  string
	EventType string
	Status    string
}

type UserLoginRepository interface {
	Create(log *models.UserLoginLog) error
	ListAdmin(filter UserLoginFilter) ([]models.UserLoginLog, int64, error)
	ListByUser(userID uint, page, pageSize int) ([]models.UserLoginLog, int64, error)
}

type AuthzRepository interface {
	Create(log *models.AuthzAuditLog) error
	ListAdmin(filter AuthzFilter) ([]models.AuthzAuditLog, int64, error)
}

type AdminLoginRepository interface {
	Create(log *models.AdminLoginLog) error
	List(filter AdminLoginFilter) ([]models.AdminLoginLog, int64, error)
}
