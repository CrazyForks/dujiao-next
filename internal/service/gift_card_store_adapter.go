package service

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/giftcard"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/repository"
)

type giftCardUserDirectoryAdapter struct {
	repo repository.UserRepository
}

func (a giftCardUserDirectoryAdapter) ListByIDs(ids []uint) ([]models.User, error) {
	if a.repo == nil {
		return nil, nil
	}
	return a.repo.ListByIDs(ids)
}

type giftCardCurrencyAdapter struct {
	settings *settingsapp.Service
}

func (a giftCardCurrencyAdapter) SiteCurrency() string {
	return resolveServiceSiteCurrency(a.settings)
}

func newGiftCardAdminService(
	repo giftcard.Repository,
	userRepo repository.UserRepository,
	settingSvc *settingsapp.Service,
) *giftcard.Service {
	return giftcard.NewService(giftcard.Options{
		Repo:     repo,
		Users:    giftCardUserDirectoryAdapter{repo: userRepo},
		Currency: giftCardCurrencyAdapter{settings: settingSvc},
	})
}
