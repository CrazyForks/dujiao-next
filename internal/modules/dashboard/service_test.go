package dashboard

import (
	"context"
	"testing"
	"time"
)

type dashboardServiceRepoStub struct {
	overview OverviewRow
	stock    StockStatsRow
}

func (s dashboardServiceRepoStub) GetOverview(startAt, endAt time.Time) (OverviewRow, error) {
	return s.overview, nil
}

func (s dashboardServiceRepoStub) GetPaymentOrderAlertCounts(startAt, endAt time.Time) (PaymentOrderAlertCountsRow, error) {
	return PaymentOrderAlertCountsRow{}, nil
}

func (s dashboardServiceRepoStub) GetOrderTrends(startAt, endAt time.Time) ([]OrderTrendRow, error) {
	return []OrderTrendRow{}, nil
}

func (s dashboardServiceRepoStub) GetPaymentTrends(startAt, endAt time.Time) ([]PaymentTrendRow, error) {
	return []PaymentTrendRow{}, nil
}

func (s dashboardServiceRepoStub) GetStockStats(lowStockThreshold int64) (StockStatsRow, error) {
	return s.stock, nil
}

func (s dashboardServiceRepoStub) GetInventoryAlertItems(lowStockThreshold int64) ([]InventoryAlertRow, error) {
	return []InventoryAlertRow{}, nil
}

func (s dashboardServiceRepoStub) GetTopProducts(startAt, endAt time.Time, limit int) ([]ProductRankingRow, error) {
	return []ProductRankingRow{}, nil
}

func (s dashboardServiceRepoStub) GetProfitOverview(startAt, endAt time.Time) (ProfitOverviewRow, error) {
	return ProfitOverviewRow{}, nil
}

func (s dashboardServiceRepoStub) GetProfitTrends(startAt, endAt time.Time) ([]ProfitTrendRow, error) {
	return []ProfitTrendRow{}, nil
}

func (s dashboardServiceRepoStub) GetTopChannels(startAt, endAt time.Time, limit int) ([]ChannelRankingRow, error) {
	return []ChannelRankingRow{}, nil
}

func (s dashboardServiceRepoStub) GetTotalUserBalance() (float64, error) {
	return 0, nil
}

func TestDashboardOverviewUsesPaidOrdersForPaymentConversionRate(t *testing.T) {
	service := NewService(dashboardServiceRepoStub{
		overview: OverviewRow{
			OrdersTotal:     10,
			PaidOrders:      6,
			CompletedOrders: 3,
			PaymentsTotal:   5,
			PaymentsSuccess: 4,
			Currency:        "cny",
			GMVPaid:         120,
		},
		stock: StockStatsRow{},
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
		overview: OverviewRow{
			PendingPaymentOrders: 25,
			PaymentsFailed:       12,
		},
		stock: StockStatsRow{
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

var _ Repository = dashboardServiceRepoStub{}
