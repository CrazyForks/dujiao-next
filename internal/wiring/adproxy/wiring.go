package adproxywiring

import (
	"context"
	"encoding/json"

	"github.com/dujiao-next/internal/modules/adproxy"
)

// adProxyAdminAdapter 将 legacy AdProxyService 适配为 transport 端口。
type adProxyAdminAdapter struct {
	svc *adproxy.Service
}

func (a adProxyAdminAdapter) RenderSlot(ctx context.Context, slotCode string, params map[string]string) (*adproxy.RenderResponse, error) {
	return a.svc.RenderSlot(ctx, slotCode, params)
}

func (a adProxyAdminAdapter) ReportImpression(ctx context.Context, payload json.RawMessage) error {
	return a.svc.ReportImpression(ctx, payload)
}
