package adminuserwiring

import (
	"context"
	"errors"
	"fmt"

	usercontract "github.com/dujiao-next/internal/modules/identity/user/contract"

	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/models"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/dujiao-next/internal/modules/coupon"
	externalidentitycontract "github.com/dujiao-next/internal/modules/identity/externalidentity/contract"
	externalidentitydomain "github.com/dujiao-next/internal/modules/identity/externalidentity/domain"
	userauthapp "github.com/dujiao-next/internal/modules/identity/userauth/application"
	"github.com/dujiao-next/internal/service"
	"github.com/dujiao-next/internal/shared/money"
	adminusertransport "github.com/dujiao-next/internal/transport/http/adminuser"
)

type adminUserDirectoryAdapter struct {
	users usercontract.Store
}

func (a adminUserDirectoryAdapter) List(filter adminusertransport.UserListFilter) ([]userdomain.User, int64, error) {
	return a.users.List(usercontract.ListFilter{
		Page:          filter.Page,
		PageSize:      filter.PageSize,
		UserID:        filter.UserID,
		Keyword:       filter.Keyword,
		Status:        filter.Status,
		CreatedFrom:   filter.CreatedFrom,
		CreatedTo:     filter.CreatedTo,
		LastLoginFrom: filter.LastLoginFrom,
		LastLoginTo:   filter.LastLoginTo,
		SortBy:        filter.SortBy,
		SortOrder:     filter.SortOrder,
	})
}

func (a adminUserDirectoryAdapter) GetByID(id uint) (*userdomain.User, error) {
	return a.users.GetByID(id)
}

func (a adminUserDirectoryAdapter) GetByEmail(email string) (*userdomain.User, error) {
	return a.users.GetByEmail(email)
}

func (a adminUserDirectoryAdapter) Update(user *userdomain.User) error {
	return a.users.Update(user)
}

func (a adminUserDirectoryAdapter) BatchUpdateStatus(ids []uint, status string) error {
	return a.users.BatchUpdateStatus(ids, status)
}

type adminUserEmailAdapter struct{}

func (adminUserEmailAdapter) NormalizeEmail(email string) (string, error) {
	normalized, err := userauthapp.NormalizeEmail(email)
	if err != nil {
		return "", mapAdminUserTransportError(err)
	}
	return normalized, nil
}

type adminUserWalletAdapter struct {
	wallets *service.WalletService
}

func (a adminUserWalletAdapter) GetBalancesByUserIDs(userIDs []uint) (map[uint]money.Amount, error) {
	return a.wallets.GetBalancesByUserIDs(userIDs)
}

func (a adminUserWalletAdapter) GetAccount(userID uint) (*models.WalletAccount, error) {
	return a.wallets.GetAccount(userID)
}

type adminUserOAuthAdapter struct {
	identities externalidentitycontract.Store
}

func (a adminUserOAuthAdapter) ListByUserID(userID uint) ([]externalidentitydomain.Identity, error) {
	return a.identities.ListByUserID(userID)
}

type adminUserTelegramAdapter struct {
	auth *userauthapp.Service
}

func (a adminUserTelegramAdapter) UnbindTelegram(userID uint) error {
	return mapAdminUserTransportError(a.auth.UnbindTelegram(userID))
}

type adminUserCouponUsageAdapter struct {
	usages coupon.UsageRepository
}

func (a adminUserCouponUsageAdapter) ListByUser(filter coupon.UsageListFilter) ([]models.CouponUsage, int64, error) {
	return a.usages.ListByUser(filter)
}

type adminUserCouponAdapter struct {
	coupons coupon.Repository
}

func (a adminUserCouponAdapter) ListByIDs(ids []uint) ([]models.Coupon, error) {
	return a.coupons.ListByIDs(ids)
}

type adminUserProductAdapter struct {
	products catalogproduct.Repository
}

func (a adminUserProductAdapter) ListByIDs(ids []uint) ([]models.Product, error) {
	return a.products.ListByIDs(ids)
}

type adminUserAuthStateAdapter struct{}

func (adminUserAuthStateAdapter) SetUserAuthState(ctx context.Context, user *userdomain.User) error {
	return cache.SetUserAuthState(ctx, cache.BuildUserAuthState(user))
}

func (adminUserAuthStateAdapter) DelUserAuthState(ctx context.Context, userID uint) error {
	return cache.DelUserAuthState(ctx, userID)
}

func mapAdminUserTransportError(err error) error {
	if err == nil {
		return nil
	}
	for _, mapping := range []struct {
		source error
		target error
	}{
		{service.ErrNotFound, adminusertransport.ErrNotFound},
		{userauthapp.ErrUserDisabled, adminusertransport.ErrUserDisabled},
		{userauthapp.ErrUserOAuthNotBound, adminusertransport.ErrUserOAuthNotBound},
		{userauthapp.ErrTelegramUnbindRequiresEmail, adminusertransport.ErrTelegramUnbindRequiresEmail},
		{userauthapp.ErrInvalidEmail, adminusertransport.ErrInvalidEmail},
	} {
		if errors.Is(err, mapping.source) {
			return fmt.Errorf("%w: %v", mapping.target, err)
		}
	}
	return err
}
