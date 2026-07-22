package dashboardhttp

import (
	"testing"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/dashboard"
)

func TestMapInventoryAlertsPreservesLocalizedAndSKUFields(t *testing.T) {
	items := mapInventoryAlerts([]dashboard.InventoryAlertRow{{
		ProductID:         7,
		SKUID:             9,
		ProductTitleJSON:  models.JSON{"zh-CN": "商品"},
		SKUCode:           "SKU-9",
		SKUSpecValuesJSON: models.JSON{"size": "L"},
		FulfillmentType:   "auto",
		AlertType:         "low_stock_products",
		AvailableStock:    2,
	}})
	if len(items) != 1 || items[0].ProductID != 7 || items[0].SKUID != 9 || items[0].SKUSpecValues["size"] != "L" {
		t.Fatalf("unexpected mapping: %+v", items)
	}
}
