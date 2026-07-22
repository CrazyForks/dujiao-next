package apicredential

import "github.com/dujiao-next/internal/models"

type ListFilter struct {
	Status   string
	UserID   uint
	Search   string
	Page     int
	PageSize int
}

type Repository interface {
	GetByID(id uint) (*models.ApiCredential, error)
	GetByUserID(userID uint) (*models.ApiCredential, error)
	GetAnyByUserID(userID uint) (*models.ApiCredential, error)
	GetByApiKey(apiKey string) (*models.ApiCredential, error)
	Create(credential *models.ApiCredential) error
	Update(credential *models.ApiCredential) error
	UpdateAny(credential *models.ApiCredential) error
	Delete(id uint) error
	List(filter ListFilter) ([]models.ApiCredential, int64, error)
}
