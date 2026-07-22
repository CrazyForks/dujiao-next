package promotion

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"

	"github.com/shopspring/decimal"
)

// AdminService 活动价管理服务。
type AdminService struct {
	repo Repository
}

// NewAdminService 创建活动价管理服务。
func NewAdminService(repo Repository) *AdminService {
	return &AdminService{repo: repo}
}

// CreatePromotionInput 创建活动价输入
type CreatePromotionInput struct {
	Name       string
	Type       string
	ScopeRefID uint
	Value      models.Money
	MinAmount  models.Money
	StartsAt   *time.Time
	EndsAt     *time.Time
	IsActive   *bool
}

// UpdatePromotionInput 更新活动价输入
type UpdatePromotionInput struct {
	Name       string
	Type       string
	ScopeRefID uint
	Value      models.Money
	MinAmount  models.Money
	StartsAt   *time.Time
	EndsAt     *time.Time
	IsActive   *bool
}

// Create 创建活动价
func (s *AdminService) Create(input CreatePromotionInput) (*models.Promotion, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrInvalid
	}
	if input.ScopeRefID == 0 {
		return nil, ErrInvalid
	}
	promotionType := strings.ToLower(strings.TrimSpace(input.Type))
	if promotionType != constants.PromotionTypeFixed && promotionType != constants.PromotionTypePercent && promotionType != constants.PromotionTypeSpecialPrice {
		return nil, ErrInvalid
	}
	if input.Value.Decimal.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalid
	}
	if promotionType == constants.PromotionTypePercent && input.Value.Decimal.GreaterThan(decimal.NewFromInt(100)) {
		return nil, ErrInvalid
	}
	if input.StartsAt != nil && input.EndsAt != nil && input.EndsAt.Before(*input.StartsAt) {
		return nil, ErrInvalid
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	promotion := &models.Promotion{
		Name:       name,
		ScopeType:  constants.ScopeTypeProduct,
		ScopeRefID: input.ScopeRefID,
		Type:       promotionType,
		Value:      input.Value,
		MinAmount:  input.MinAmount,
		StartsAt:   input.StartsAt,
		EndsAt:     input.EndsAt,
		IsActive:   isActive,
	}

	if err := s.repo.Create(promotion); err != nil {
		return nil, err
	}
	return promotion, nil
}

// Update 更新活动价
func (s *AdminService) Update(id uint, input UpdatePromotionInput) (*models.Promotion, error) {
	if id == 0 {
		return nil, ErrInvalid
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrInvalid
	}
	if input.ScopeRefID == 0 {
		return nil, ErrInvalid
	}
	promotionType := strings.ToLower(strings.TrimSpace(input.Type))
	if promotionType != constants.PromotionTypeFixed && promotionType != constants.PromotionTypePercent && promotionType != constants.PromotionTypeSpecialPrice {
		return nil, ErrInvalid
	}
	if input.Value.Decimal.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalid
	}
	if promotionType == constants.PromotionTypePercent && input.Value.Decimal.GreaterThan(decimal.NewFromInt(100)) {
		return nil, ErrInvalid
	}
	if input.StartsAt != nil && input.EndsAt != nil && input.EndsAt.Before(*input.StartsAt) {
		return nil, ErrInvalid
	}

	isActive := existing.IsActive
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	existing.Name = name
	existing.ScopeType = constants.ScopeTypeProduct
	existing.ScopeRefID = input.ScopeRefID
	existing.Type = promotionType
	existing.Value = input.Value
	existing.MinAmount = input.MinAmount
	existing.StartsAt = input.StartsAt
	existing.EndsAt = input.EndsAt
	existing.IsActive = isActive

	if err := s.repo.Update(existing); err != nil {
		return nil, ErrUpdateFailed
	}
	return existing, nil
}

// Delete 删除活动价
func (s *AdminService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalid
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.repo.Delete(id); err != nil {
		return ErrDeleteFailed
	}
	return nil
}

// List 获取活动价列表
func (s *AdminService) List(filter ListFilter) ([]models.Promotion, int64, error) {
	return s.repo.List(filter)
}
