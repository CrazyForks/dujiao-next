package dashboardhttp

import (
	"context"
	"errors"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/modules/dashboard"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

// Reader is the minimal dashboard use-case surface consumed by HTTP.
type Reader interface {
	GetOverview(ctx context.Context, input dashboard.QueryInput) (*dashboard.OverviewResponse, error)
	GetTrends(ctx context.Context, input dashboard.QueryInput) (*dashboard.TrendResponse, error)
	GetRankings(ctx context.Context, input dashboard.QueryInput) (*dashboard.RankingsResponse, error)
	LoadDashboardAlertSetting() settingsmodule.DashboardAlertSetting
	GetInventoryAlertItems(ctx context.Context, lowStockThreshold int64) ([]dashboard.InventoryAlertRow, error)
}

type AdminHandler struct {
	reader Reader
}

func NewAdminHandler(reader Reader) *AdminHandler {
	return &AdminHandler{reader: reader}
}

func (h *AdminHandler) GetOverview(c *gin.Context) {
	input, ok := parseQuery(c)
	if !ok {
		return
	}
	if h == nil || h.reader == nil {
		respondFetchError(c, nil)
		return
	}
	data, err := h.reader.GetOverview(c.Request.Context(), input)
	if err != nil {
		respondFetchError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *AdminHandler) GetTrends(c *gin.Context) {
	input, ok := parseQuery(c)
	if !ok {
		return
	}
	if h == nil || h.reader == nil {
		respondFetchError(c, nil)
		return
	}
	data, err := h.reader.GetTrends(c.Request.Context(), input)
	if err != nil {
		respondFetchError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *AdminHandler) GetRankings(c *gin.Context) {
	input, ok := parseQuery(c)
	if !ok {
		return
	}
	if h == nil || h.reader == nil {
		respondFetchError(c, nil)
		return
	}
	data, err := h.reader.GetRankings(c.Request.Context(), input)
	if err != nil {
		respondFetchError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *AdminHandler) GetInventoryAlerts(c *gin.Context) {
	if h == nil || h.reader == nil {
		respondFetchError(c, nil)
		return
	}
	setting := h.reader.LoadDashboardAlertSetting()
	items, err := h.reader.GetInventoryAlertItems(c.Request.Context(), setting.LowStockThreshold)
	if err != nil {
		respondFetchError(c, err)
		return
	}
	response.Success(c, mapInventoryAlerts(items))
}

func parseQuery(c *gin.Context) (dashboard.QueryInput, bool) {
	input, err := shared.ParseReportingQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return dashboard.QueryInput{}, false
	}
	return input, true
}

func respondFetchError(c *gin.Context, err error) {
	if errors.Is(err, dashboard.ErrRangeInvalid) {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	shared.RespondError(c, response.CodeInternal, "error.dashboard_fetch_failed", err)
}
