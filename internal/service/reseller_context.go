package service

import (
	"context"
	"net/http"

	"github.com/dujiao-next/internal/config"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
)

// TenantContext 保留旧 Service API；真实值对象位于 modules/reseller。
type TenantContext = resellermodule.TenantContext

func MainTenantContext(host string) TenantContext {
	return resellermodule.MainTenantContext(host)
}

func ResellerTenantContext(host string, resellerID uint, resellerUserID uint, primaryDomain string) TenantContext {
	return resellermodule.ResellerTenantContext(host, resellerID, resellerUserID, primaryDomain)
}

func UnavailableTenantContext(host string, reason string) TenantContext {
	return resellermodule.UnavailableTenantContext(host, reason)
}

func WithTenantContext(ctx context.Context, tenant TenantContext) context.Context {
	return resellermodule.WithTenantContext(ctx, tenant)
}

func TenantFromContext(ctx context.Context) (TenantContext, bool) {
	return resellermodule.TenantFromContext(ctx)
}

func NormalizeResellerHost(raw string) string {
	return resellermodule.NormalizeHost(raw)
}

func ResolveResellerRequestHost(req *http.Request, cfg config.ResellerConfig) string {
	return resellermodule.ResolveRequestHost(req, cfg)
}
