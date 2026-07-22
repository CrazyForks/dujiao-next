package cardsecret

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrInsufficient       = errors.New("card secret insufficient")
	ErrInvalid            = errors.New("card secret invalid")
	ErrCreateFailed       = errors.New("card secret create failed")
	ErrFetchFailed        = errors.New("card secret fetch failed")
	ErrUpdateFailed       = errors.New("card secret update failed")
	ErrDeleteFailed       = errors.New("card secret delete failed")
	ErrBatchCreateFailed  = errors.New("card secret batch create failed")
	ErrBatchFetchFailed   = errors.New("card secret batch fetch failed")
	ErrImportFailed       = errors.New("card secret import failed")
	ErrStatsFailed        = errors.New("card secret stats failed")
	ErrProductFetchFailed = errors.New("product fetch failed")
	ErrProductNotFound    = errors.New("product not found")
	ErrProductSKURequired = errors.New("product sku required")
	ErrProductSKUInvalid  = errors.New("product sku invalid")
)

type ListFilter struct {
	ProductID uint
	SKUID     uint
	BatchID   uint
	Status    string
	Secret    string
	BatchNo   string
	Page      int
	PageSize  int
}

type BatchStatusCount struct {
	BatchID uint   `gorm:"column:batch_id"`
	Status  string `gorm:"column:status"`
	Total   int64  `gorm:"column:total"`
}

type SKUStockCount struct {
	ProductID uint   `gorm:"column:product_id"`
	SKUID     uint   `gorm:"column:sku_id"`
	Status    string `gorm:"column:status"`
	Total     int64  `gorm:"column:total"`
}

type Repository interface {
	CreateBatch(items []models.CardSecret) error
	List(filter ListFilter) ([]models.CardSecret, int64, error)
	ListIDs(filter ListFilter) ([]uint, error)
	ListByIDs(ids []uint) ([]models.CardSecret, error)
	ListIDsByBatchID(batchID uint) ([]uint, error)
	CountByBatchIDs(batchIDs []uint) ([]BatchStatusCount, error)
	ListByOrderAndStatus(orderID uint, status string) ([]models.CardSecret, error)
	ListAvailableByProduct(productID, skuID uint, limit int) ([]models.CardSecret, error)
	ListAvailableByProductForUpdate(productID, skuID uint, limit int) ([]models.CardSecret, error)
	ListAvailableByProductBatchForUpdate(productID, skuID, batchID uint, limit int) ([]models.CardSecret, error)
	GetByID(id uint) (*models.CardSecret, error)
	Update(secret *models.CardSecret) error
	BatchUpdateStatus(ids []uint, status string, updatedAt time.Time) (int64, error)
	BatchDeleteByIDs(ids []uint) (int64, error)
	CountByProduct(productID, skuID uint) (int64, int64, int64, error)
	CountAvailable(productID, skuID uint) (int64, error)
	CountAvailableByProductIDs(productIDs []uint) (map[uint]int64, error)
	CountReserved(productID, skuID uint) (int64, error)
	CountStockByProductIDs(productIDs []uint) ([]SKUStockCount, error)
	Reserve(ids []uint, orderID uint, reservedAt time.Time) (int64, error)
	ReleaseByOrder(orderID uint) (int64, error)
	MarkUsed(ids []uint, orderID uint, usedAt time.Time) (int64, error)
	DeleteByProduct(productID uint) error
}

type BatchRepository interface {
	Create(batch *models.CardSecretBatch) error
	GetByID(id uint) (*models.CardSecretBatch, error)
	ListByProduct(productID, skuID uint, page, pageSize int) ([]models.CardSecretBatch, int64, error)
	DeleteByProduct(productID uint) error
}

type UnitOfWork interface {
	Transaction(fn func(secrets Repository, batches BatchRepository) error) error
}

type ProductRepository interface {
	GetByID(id string) (*models.Product, error)
}

type ProductSKURepository interface {
	ListByProduct(productID uint, includeInactive bool) ([]models.ProductSKU, error)
	GetByID(id uint) (*models.ProductSKU, error)
	GetByProductAndCode(productID uint, skuCode string) (*models.ProductSKU, error)
}

type ServiceOptions struct {
	Secrets      Repository
	Batches      BatchRepository
	Transactions UnitOfWork
	Products     ProductRepository
	ProductSKUs  ProductSKURepository
}

// Service 卡密库存服务。
type Service struct {
	secretRepo     Repository
	batchRepo      BatchRepository
	transactions   UnitOfWork
	productRepo    ProductRepository
	productSKURepo ProductSKURepository
}

func NewService(options ServiceOptions) *Service {
	return &Service{
		secretRepo:     options.Secrets,
		batchRepo:      options.Batches,
		transactions:   options.Transactions,
		productRepo:    options.Products,
		productSKURepo: options.ProductSKUs,
	}
}

func (s *Service) resolveCardSecretSKU(productID, rawSKUID uint) (*models.ProductSKU, error) {
	if productID == 0 || s.productSKURepo == nil {
		return nil, ErrProductSKUInvalid
	}
	product, err := s.productRepo.GetByID(strings.TrimSpace(strconv.FormatUint(uint64(productID), 10)))
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	skus, err := s.productSKURepo.ListByProduct(productID, false)
	if err != nil {
		return nil, err
	}
	activeSKUs := make([]models.ProductSKU, 0, len(skus))
	for _, sku := range skus {
		if !sku.IsActive {
			continue
		}
		activeSKUs = append(activeSKUs, sku)
	}
	if rawSKUID > 0 {
		sku, err := s.productSKURepo.GetByID(rawSKUID)
		if err != nil {
			return nil, err
		}
		if sku == nil || sku.ProductID != productID {
			return nil, ErrProductSKUInvalid
		}
		if strings.TrimSpace(product.FulfillmentType) == constants.FulfillmentTypeAuto && !sku.IsActive {
			return nil, ErrProductSKUInvalid
		}
		return sku, nil
	}

	if strings.TrimSpace(product.FulfillmentType) == constants.FulfillmentTypeAuto {
		switch len(activeSKUs) {
		case 0:
		case 1:
			return &activeSKUs[0], nil
		default:
			return nil, ErrProductSKURequired
		}
	}

	defaultSKU, err := s.productSKURepo.GetByProductAndCode(productID, models.DefaultSKUCode)
	if err != nil {
		return nil, err
	}
	if defaultSKU != nil {
		return defaultSKU, nil
	}
	if len(skus) == 1 {
		return &skus[0], nil
	}
	return nil, ErrProductSKURequired
}

func normalizeCardSecretIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return []uint{}
	}
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
