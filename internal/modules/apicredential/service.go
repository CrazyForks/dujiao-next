package apicredential

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

var (
	ErrApiCredentialExists       = errors.New("api credential already exists for this user")
	ErrApiCredentialNotFound     = errors.New("api credential not found")
	ErrApiCredentialNotApproved  = errors.New("api credential is not approved")
	ErrApiCredentialPendingExist = errors.New("pending application already exists")
)

// Service API 凭证服务
type Service struct {
	credRepo Repository
}

// NewService 创建凭证服务
func NewService(credRepo Repository) *Service {
	return &Service{credRepo: credRepo}
}

// Apply 用户申请 API 对接权限
func (s *Service) Apply(userID uint) (*models.ApiCredential, error) {
	existing, err := s.credRepo.GetAnyByUserID(userID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		if existing.DeletedAt.Valid {
			if err := resetApiCredentialForReapply(existing); err != nil {
				return nil, err
			}
			if err := s.credRepo.UpdateAny(existing); err != nil {
				return nil, err
			}
			return existing, nil
		}

		switch existing.Status {
		case constants.ApiCredentialStatusPendingReview:
			return nil, ErrApiCredentialPendingExist
		case constants.ApiCredentialStatusApproved:
			return nil, ErrApiCredentialExists
		case constants.ApiCredentialStatusRejected:
			// 允许重新申请，并重置旧审批与凭证痕迹。
			if err := resetApiCredentialForReapply(existing); err != nil {
				return nil, err
			}
			if err := s.credRepo.Update(existing); err != nil {
				return nil, err
			}
			return existing, nil
		case constants.ApiCredentialStatusDisabled:
			return nil, ErrApiCredentialExists
		}
	}

	apiKey, err := generateRandomHex(32)
	if err != nil {
		return nil, err
	}
	cred := &models.ApiCredential{
		UserID: userID,
		ApiKey: apiKey,
		Status: constants.ApiCredentialStatusPendingReview,
	}
	if err := s.credRepo.Create(cred); err != nil {
		return nil, err
	}
	return cred, nil
}

func resetApiCredentialForReapply(cred *models.ApiCredential) error {
	apiKey, err := generateRandomHex(32)
	if err != nil {
		return err
	}
	cred.ApiKey = apiKey
	cred.ApiSecret = ""
	cred.Status = constants.ApiCredentialStatusPendingReview
	cred.RejectReason = ""
	cred.ApprovedAt = nil
	cred.LastUsedAt = nil
	cred.IsActive = false
	cred.DeletedAt = models.ApiCredential{}.DeletedAt
	return nil
}

// Approve admin 审核通过
func (s *Service) Approve(id uint) (*models.ApiCredential, string, error) {
	cred, err := s.credRepo.GetByID(id)
	if err != nil {
		return nil, "", err
	}
	if cred == nil {
		return nil, "", ErrApiCredentialNotFound
	}

	apiKey, err := generateRandomHex(32)
	if err != nil {
		return nil, "", err
	}
	apiSecret, err := generateRandomHex(64)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	cred.ApiKey = apiKey
	cred.ApiSecret = apiSecret
	cred.Status = constants.ApiCredentialStatusApproved
	cred.ApprovedAt = &now
	cred.IsActive = true
	cred.RejectReason = ""

	if err := s.credRepo.Update(cred); err != nil {
		return nil, "", err
	}

	return cred, apiSecret, nil
}

// Reject admin 审核拒绝
func (s *Service) Reject(id uint, reason string) error {
	cred, err := s.credRepo.GetByID(id)
	if err != nil {
		return err
	}
	if cred == nil {
		return ErrApiCredentialNotFound
	}

	cred.Status = constants.ApiCredentialStatusRejected
	cred.RejectReason = reason
	return s.credRepo.Update(cred)
}

// SetActive 启用/禁用
func (s *Service) SetActive(id uint, active bool) error {
	cred, err := s.credRepo.GetByID(id)
	if err != nil {
		return err
	}
	if cred == nil {
		return ErrApiCredentialNotFound
	}
	if cred.Status != constants.ApiCredentialStatusApproved {
		return ErrApiCredentialNotApproved
	}

	cred.IsActive = active
	return s.credRepo.Update(cred)
}

// SetActiveByUserID 用户自行启用/禁用
func (s *Service) SetActiveByUserID(userID uint, active bool) error {
	cred, err := s.credRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if cred == nil {
		return ErrApiCredentialNotFound
	}
	if cred.Status != constants.ApiCredentialStatusApproved {
		return ErrApiCredentialNotApproved
	}

	cred.IsActive = active
	return s.credRepo.Update(cred)
}

// Regenerate 重新生成 Secret
func (s *Service) Regenerate(id uint) (string, error) {
	cred, err := s.credRepo.GetByID(id)
	if err != nil {
		return "", err
	}
	if cred == nil {
		return "", ErrApiCredentialNotFound
	}
	if cred.Status != constants.ApiCredentialStatusApproved {
		return "", ErrApiCredentialNotApproved
	}

	newSecret, err := generateRandomHex(64)
	if err != nil {
		return "", err
	}

	cred.ApiSecret = newSecret
	if err := s.credRepo.Update(cred); err != nil {
		return "", err
	}
	return newSecret, nil
}

// RegenerateByUserID 用户重新生成 Secret
func (s *Service) RegenerateByUserID(userID uint) (string, error) {
	cred, err := s.credRepo.GetByUserID(userID)
	if err != nil {
		return "", err
	}
	if cred == nil {
		return "", ErrApiCredentialNotFound
	}
	return s.Regenerate(cred.ID)
}

// GetByUserID 获取用户的凭证
func (s *Service) GetByUserID(userID uint) (*models.ApiCredential, error) {
	return s.credRepo.GetByUserID(userID)
}

// GetByID 根据 ID 获取凭证
func (s *Service) GetByID(id uint) (*models.ApiCredential, error) {
	return s.credRepo.GetByID(id)
}

// List 列表查询
func (s *Service) List(filter ListFilter) ([]models.ApiCredential, int64, error) {
	return s.credRepo.List(filter)
}

// Delete 删除凭证
func (s *Service) Delete(id uint) error {
	return s.credRepo.Delete(id)
}

func generateRandomHex(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
