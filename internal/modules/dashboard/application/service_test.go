package application

import (
	"context"
	"testing"
	"time"

	dashboardcontract "github.com/dujiao-next/internal/modules/dashboard/contract"
)

type dashboardServiceRepoStub struct {
	overview dashboardcontract.OverviewRow
	stock    dashboardcontract.StockStatsRow
}

func (s dashboardServiceRepoStub) GetOverview(startAt, endAt time.Time) (dashboardcontract.OverviewRow, error) {
	return s.overview, nil
}

func (s dashboardServiceRepoStub) GetPaymentOrderAlertCounts(startAt, endAt time.Time) (dashboardcontract.PaymentOrderAlertCountsRow, error) {
	return dashboardcontract.PaymentOrderAlertCountsRow{}, nil
}

func (s dashboardServiceRepoStub) GetOrderTrends(startAt, endAt time.Time) ([]dashboardcontract.OrderTrendRow, error) {
	return []dashboardcontract.OrderTrendRow{}, nil
}

func (s dashboardServiceRepoStub) GetPaymentTrends(startAt, endAt time.Time) ([]dashboardcontract.PaymentTrendRow, error) {
	return []dashboardcontract.PaymentTrendRow{}, nil
}

func (s dashboardServiceRepoStub) GetStockStats(lowStockThreshold int64) (dashboardcontract.StockStatsRow, error) {
	return s.stock, nil
}

func (s dashboardServiceRepoStub) GetInventoryAlertItems(lowStockThreshold int64) ([]dashboardcontract.InventoryAlertRow, error) {
	return []dashboardcontract.InventoryAlertRow{}, nil
}

func (s dashboardServiceRepoStub) GetTopProducts(startAt, endAt time.Time, limit int) ([]dashboardcontract.ProductRankingRow, error) {
	return []dashboardcontract.ProductRankingRow{}, nil
}

func (s dashboardServiceRepoStub) GetProfitOverview(startAt, endAt time.Time) (dashboardcontract.ProfitOverviewRow, error) {
	return dashboardcontract.ProfitOverviewRow{}, nil
}

func (s dashboardServiceRepoStub) GetProfitTrends(startAt, endAt time.Time) ([]dashboardcontract.ProfitTrendRow, error) {
	return []dashboardcontract.ProfitTrendRow{}, nil
}

func (s dashboardServiceRepoStub) GetTopChannels(startAt, endAt time.Time, limit int) ([]dashboardcontract.ChannelRankingRow, error) {
	return []dashboardcontract.ChannelRankingRow{}, nil
}

func (s dashboardServiceRepoStub) GetTotalUserBalance() (float64, error) {
	return 0, nil
}

func TestDashboardOverviewUsesPaidOrdersForPaymentConversionRate(t *testing.T) {
	service := NewService(dashboardServiceRepoStub{
		overview: dashboardcontract.OverviewRow{
			OrdersTotal:     10,
			PaidOrders:      6,
			CompletedOrders: 3,
			PaymentsTotal:   5,
			PaymentsSuccess: 4,
			Currency:        "cny",
			GMVPaid:         120,
		},
		stock: dashboardcontract.StockStatsRow{},
	}, nil)

	response, err := service.GetOverview(context.Background(), QueryInput{
		Range:    "today",
		Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("get overview failed: %v", err)
	}
	if response.Currency != "CNY" {
		t.Fatalf("currency want CNY got %s", response.Currency)
	}
	if response.Funnel.PaymentConversionRate != "60.00" {
		t.Fatalf("payment conversion rate want 60.00 got %s", response.Funnel.PaymentConversionRate)
	}
	if response.KPI.PaymentSuccessRate != "80.00" {
		t.Fatalf("payment success rate want 80.00 got %s", response.KPI.PaymentSuccessRate)
	}
}

func TestDashboardOverviewBuildsInventoryAlertsFromStockStats(t *testing.T) {
	service := NewService(dashboardServiceRepoStub{
		overview: dashboardcontract.OverviewRow{
			PendingPaymentOrders: 25,
			PaymentsFailed:       12,
		},
		stock: dashboardcontract.StockStatsRow{
			OutOfStockProducts: 2,
			LowStockProducts:   1,
		},
	}, nil)

	response, err := service.GetOverview(context.Background(), QueryInput{
		Range:    "today",
		Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("get overview failed: %v", err)
	}
	if len(response.Alerts) != 4 {
		t.Fatalf("alerts len want 4 got %d", len(response.Alerts))
	}
	if response.Alerts[0].Type != "out_of_stock_products" || response.Alerts[0].Value != 2 {
		t.Fatalf("unexpected first alert: %+v", response.Alerts[0])
	}
}

var _ dashboardcontract.Repository = dashboardServiceRepoStub{}
