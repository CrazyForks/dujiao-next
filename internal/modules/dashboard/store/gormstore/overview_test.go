package gormstore

import (
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"

	"github.com/shopspring/decimal"
)

func TestGetOverviewUsesOrderCreationWindowForPaidGMV(t *testing.T) {
	repo, db := setupDashboardRepositoryTest(t)
	now := time.Now().UTC().Truncate(time.Second)

	paidOutsideWindow := now.Add(24 * time.Hour)
	inWindowOrder := &models.Order{
		OrderNo:        "DJ-GMV-IN-WINDOW",
		UserID:         1,
		Status:         constants.OrderStatusPaid,
		Currency:       "CNY",
		OriginalAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		DiscountAmount: models.NewMoneyFromDecimal(decimal.Zero),
		TotalAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		CreatedAt:      now,
		PaidAt:         &paidOutsideWindow,
	}
	if err := db.Create(inWindowOrder).Error; err != nil {
		t.Fatalf("create in-window order failed: %v", err)
	}

	paidInsideWindow := now
	outOfWindowOrder := &models.Order{
		OrderNo:        "DJ-GMV-OUT-WINDOW",
		UserID:         1,
		Status:         constants.OrderStatusPaid,
		Currency:       "CNY",
		OriginalAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(60)),
		DiscountAmount: models.NewMoneyFromDecimal(decimal.Zero),
		TotalAmount:    models.NewMoneyFromDecimal(decimal.NewFromInt(60)),
		CreatedAt:      now.Add(-48 * time.Hour),
		PaidAt:         &paidInsideWindow,
	}
	if err := db.Create(outOfWindowOrder).Error; err != nil {
		t.Fatalf("create out-of-window order failed: %v", err)
	}

	overview, err := repo.GetOverview(now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("get overview failed: %v", err)
	}
	if overview.GMVPaid != 100 {
		t.Fatalf("gmv paid want 100 got %.2f", overview.GMVPaid)
	}
}
