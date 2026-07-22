package giftcard

import (
	"strings"
	"time"

	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"

	"github.com/dujiao-next/internal/models"
)

// List 获取礼品卡列表。
func (s *Service) List(input ListInput) ([]models.GiftCard, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, ErrFetchFailed
	}
	cards, total, err := s.repo.List(ListFilter{
		Code:           strings.TrimSpace(strings.ToUpper(input.Code)),
		Status:         strings.TrimSpace(strings.ToLower(input.Status)),
		BatchNo:        strings.TrimSpace(strings.ToUpper(input.BatchNo)),
		RedeemedUserID: input.RedeemedUserID,
		CreatedFrom:    input.CreatedFrom,
		CreatedTo:      input.CreatedTo,
		RedeemedFrom:   input.RedeemedFrom,
		RedeemedTo:     input.RedeemedTo,
		ExpiresFrom:    input.ExpiresFrom,
		ExpiresTo:      input.ExpiresTo,
		Page:           input.Page,
		PageSize:       input.PageSize,
	})
	if err != nil {
		return nil, 0, ErrFetchFailed
	}
	return cards, total, nil
}

// Update 更新礼品卡。
func (s *Service) Update(id uint, input UpdateInput) (*models.GiftCard, error) {
	if s == nil || s.repo == nil || id == 0 {
		return nil, ErrInvalid
	}
	card, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrFetchFailed
	}
	if card == nil {
		return nil, ErrNotFound
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, ErrInvalid
		}
		card.Name = name
	}
	if input.Status != nil {
		status := strings.TrimSpace(strings.ToLower(*input.Status))
		switch status {
		case models.GiftCardStatusActive, models.GiftCardStatusDisabled:
			if card.Status == models.GiftCardStatusRedeemed {
				return nil, ErrInvalid
			}
			card.Status = status
		default:
			return nil, ErrInvalid
		}
	}
	if input.ClearExpiresAt {
		card.ExpiresAt = nil
	} else if input.ExpiresAt != nil {
		normalized := normalizeExpireAt(input.ExpiresAt)
		if normalized != nil && normalized.Before(time.Now()) {
			return nil, ErrInvalid
		}
		card.ExpiresAt = normalized
	}
	card.UpdatedAt = time.Now()
	if err := s.repo.Update(card); err != nil {
		return nil, ErrUpdateFailed
	}
	return card, nil
}

// Delete 删除礼品卡。
func (s *Service) Delete(id uint) error {
	if s == nil || s.repo == nil || id == 0 {
		return ErrInvalid
	}
	card, err := s.repo.GetByID(id)
	if err != nil {
		return ErrFetchFailed
	}
	if card == nil {
		return ErrNotFound
	}
	if card.Status == models.GiftCardStatusRedeemed {
		return ErrInvalid
	}
	if err := s.repo.Delete(id); err != nil {
		return ErrDeleteFailed
	}
	return nil
}

// BatchUpdateStatus 批量更新礼品卡状态。
func (s *Service) BatchUpdateStatus(ids []uint, status string) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, ErrInvalid
	}
	normalizedIDs := normalizeIDs(ids)
	if len(normalizedIDs) == 0 {
		return 0, ErrInvalid
	}
	normalizedStatus := strings.TrimSpace(strings.ToLower(status))
	switch normalizedStatus {
	case models.GiftCardStatusActive, models.GiftCardStatusDisabled:
	default:
		return 0, ErrInvalid
	}
	rows, err := s.repo.BatchUpdateStatus(normalizedIDs, normalizedStatus, time.Now())
	if err != nil {
		return 0, ErrUpdateFailed
	}
	return rows, nil
}

// ResolveRedeemedUsers 批量解析礼品卡兑换用户。
func (s *Service) ResolveRedeemedUsers(cards []models.GiftCard) (map[uint]userdomain.User, error) {
	result := make(map[uint]userdomain.User)
	if s == nil || s.users == nil || len(cards) == 0 {
		return result, nil
	}
	userIDs := make([]uint, 0, len(cards))
	seen := make(map[uint]struct{})
	for _, card := range cards {
		if card.RedeemedUserID == nil || *card.RedeemedUserID == 0 {
			continue
		}
		if _, ok := seen[*card.RedeemedUserID]; ok {
			continue
		}
		seen[*card.RedeemedUserID] = struct{}{}
		userIDs = append(userIDs, *card.RedeemedUserID)
	}
	if len(userIDs) == 0 {
		return result, nil
	}
	users, err := s.users.ListByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		result[user.ID] = user
	}
	return result, nil
}
