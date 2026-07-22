package reseller

import (
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

func TestNeutralProfitStatusUnavailableWhenIneligible(t *testing.T) {
	status := neutralProfitStatus(models.ResellerOrderSnapshot{
		ProfitEligible: false,
		ProfitAmount:   money.FromDecimal(decimal.NewFromInt(10)),
	}, models.Order{Status: constants.OrderStatusPaid}, nil)
	if status != ProfitStatusUnavailable {
		t.Fatalf("expected unavailable, got %s", status)
	}
}

func TestMaskBuyerEmail(t *testing.T) {
	if got := maskBuyerEmail("ashang@example.com"); got != "a***@example.com" {
		t.Fatalf("unexpected mask: %s", got)
	}
}

func TestOrderQueryServiceRejectsInactiveProfile(t *testing.T) {
	svc := NewOrderQueryService(orderQueryStoreStub{profile: &models.ResellerProfile{
		ID:     1,
		UserID: 9,
		Status: models.ResellerProfileStatusDisabled,
	}})
	_, _, err := svc.ListUserOrders(9, OrderListInput{Page: 1, PageSize: 10})
	if err != ErrProfileInactive {
		t.Fatalf("expected profile inactive, got %v", err)
	}
}

type orderQueryStoreStub struct {
	profile *models.ResellerProfile
}

func (s orderQueryStoreStub) GetProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	return s.profile, nil
}
func (s orderQueryStoreStub) GetProfileByID(id uint) (*models.ResellerProfile, error) {
	return s.profile, nil
}
func (s orderQueryStoreStub) ListOrderSnapshotsByReseller(filter OrderSnapshotListFilter) ([]OrderSnapshotRow, int64, error) {
	return nil, 0, nil
}
func (s orderQueryStoreStub) StatsOrderSnapshotsByReseller(filter OrderSnapshotListFilter) (OrderStatsRow, error) {
	return OrderStatsRow{}, nil
}
func (s orderQueryStoreStub) GetOrderSnapshotByResellerOrderNo(resellerID uint, orderNo string) (*OrderSnapshotRow, error) {
	return nil, nil
}
