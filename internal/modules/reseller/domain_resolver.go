package reseller

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/models"
)

const (
	// DomainUnavailableNotFound 表示域名未找到或分销站不可用。
	DomainUnavailableNotFound = "not_found"
)

// DomainResolver 将请求 Host 解析为主站或分销站上下文。
type DomainResolver struct {
	repo      DomainLookupRepository
	cfg       config.ResellerConfig
	mainHosts map[string]struct{}
}

// NewDomainResolver 创建域名解析器。
func NewDomainResolver(repo DomainLookupRepository, cfg config.ResellerConfig) *DomainResolver {
	mainHosts := make(map[string]struct{}, len(cfg.MainHosts))
	for _, host := range cfg.MainHosts {
		normalized := NormalizeHost(host)
		if normalized != "" {
			mainHosts[normalized] = struct{}{}
		}
	}
	return &DomainResolver{repo: repo, cfg: cfg, mainHosts: mainHosts}
}

// ResolveRequest 解析 HTTP 请求的租户上下文。
func (r *DomainResolver) ResolveRequest(ctx context.Context, req *http.Request) (TenantContext, error) {
	if r == nil {
		return MainTenantContext(""), nil
	}
	return r.ResolveHost(ctx, ResolveRequestHost(req, r.cfg))
}

// ResolveHost 按原始 Host 解析租户上下文。
func (r *DomainResolver) ResolveHost(ctx context.Context, rawHost string) (TenantContext, error) {
	host := NormalizeHost(rawHost)
	if r == nil || !r.cfg.Enabled {
		return MainTenantContext(host), nil
	}
	if host == "" {
		return MainTenantContext(host), nil
	}
	if _, ok := r.mainHosts[host]; ok {
		return MainTenantContext(host), nil
	}
	var cached cache.ResellerDomainCacheValue
	if hit, err := cache.GetResellerDomain(ctx, host, &cached); err == nil && hit {
		return ResellerTenantContext(host, cached.ResellerID, cached.ResellerUserID, cached.PrimaryDomain), nil
	}
	if hit, err := cache.GetResellerDomainNotFound(ctx, host); err == nil && hit {
		return UnavailableTenantContext(host, DomainUnavailableNotFound), nil
	}
	if r.repo == nil {
		return TenantContext{}, errors.New("reseller domain repository is nil")
	}
	domain, err := r.repo.FindActiveVerifiedDomain(host)
	if err != nil {
		return TenantContext{}, err
	}
	if domain == nil || domain.Profile == nil || domain.Profile.Status != models.ResellerProfileStatusActive {
		_ = cache.SetResellerDomainNotFound(ctx, host)
		return UnavailableTenantContext(host, DomainUnavailableNotFound), nil
	}
	primaryDomain := strings.TrimSpace(domain.Domain)
	value := cache.ResellerDomainCacheValue{
		ResellerID:         domain.ResellerID,
		ResellerUserID:     domain.Profile.UserID,
		Domain:             domain.Domain,
		PrimaryDomain:      primaryDomain,
		Status:             domain.Status,
		VerificationStatus: domain.VerificationStatus,
	}
	_ = cache.SetResellerDomain(ctx, host, value)
	return ResellerTenantContext(host, domain.ResellerID, domain.Profile.UserID, primaryDomain), nil
}
