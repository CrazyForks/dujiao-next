package gormstore

import (
	"errors"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/memberlevel"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// UserStore persists only the user fields owned by member-level progression.
type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (r *UserStore) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserStore) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserStore) IncrementTotalRecharged(userID uint, amount decimal.Decimal) error {
	return r.incrementMoneyColumn(userID, "total_recharged", amount)
}

func (r *UserStore) IncrementTotalSpent(userID uint, amount decimal.Decimal) error {
	return r.incrementMoneyColumn(userID, "total_spent", amount)
}

func (r *UserStore) incrementMoneyColumn(userID uint, column string, amount decimal.Decimal) error {
	if userID == 0 {
		return nil
	}
	amount = amount.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			column:       gorm.Expr(column+" + ?", models.NewMoneyFromDecimal(amount)),
			"updated_at": time.Now(),
		}).Error
}

func (r *UserStore) UpdateMemberLevelIfCurrent(userID, currentLevelID, nextLevelID uint) (int64, error) {
	if userID == 0 || currentLevelID == nextLevelID {
		return 0, nil
	}
	result := r.db.Model(&models.User{}).
		Where("id = ? AND member_level_id = ?", userID, currentLevelID).
		Updates(map[string]interface{}{
			"member_level_id": nextLevelID,
			"updated_at":      time.Now(),
		})
	return result.RowsAffected, result.Error
}

func (r *UserStore) AssignDefaultMemberLevel(defaultLevelID uint) (int64, error) {
	if defaultLevelID == 0 {
		return 0, nil
	}
	result := r.db.Model(&models.User{}).
		Where("member_level_id = 0 OR member_level_id IS NULL").
		Updates(map[string]interface{}{
			"member_level_id": defaultLevelID,
			"updated_at":      time.Now(),
		})
	return result.RowsAffected, result.Error
}

var _ memberlevel.UserRepository = (*UserStore)(nil)
