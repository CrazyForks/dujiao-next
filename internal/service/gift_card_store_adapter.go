package service

import (
	"github.com/dujiao-next/internal/modules/giftcard"
	usercontract "github.com/dujiao-next/internal/modules/identity/user/contract"
	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
)

type giftCardUserDirectoryAdapter struct {
	repo usercontract.Store
}

func (a giftCardUserDirectoryAdapter) ListByIDs(ids []uint) ([]userdomain.User, error) {
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
	userRepo usercontract.Store,
	settingSvc *settingsapp.Service,
) *giftcard.Service {
	return giftcard.NewService(giftcard.Options{
		Repo:     repo,
		Users:    giftCardUserDirectoryAdapter{repo: userRepo},
		Currency: giftCardCurrencyAdapter{settings: settingSvc},
	})
}
