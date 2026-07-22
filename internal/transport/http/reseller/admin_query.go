package resellerhttp

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

// ProfileListFilter 管理端资料列表过滤条件。
type ProfileListFilter struct {
	Page             int
	PageSize         int
	UserID           uint
	Status           string
	SettlementStatus string
	Keyword          string
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
}

// DomainListFilter 管理端域名列表过滤条件。
type DomainListFilter struct {
	Page               int
	PageSize           int
	ResellerID         uint
	UserID             uint
	Domain             string
	Type               string
	Status             string
	VerificationStatus string
	Keyword            string
	CreatedFrom        *time.Time
	CreatedTo          *time.Time
}

// SiteConfigListFilter 管理端站点配置列表过滤条件。
type SiteConfigListFilter struct {
	Page        int
	PageSize    int
	ResellerID  uint
	Keyword     string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

// ProfileDirectory 是管理端资料/域名列表与只读查询端口。
type ProfileDirectory interface {
	ListProfiles(filter ProfileListFilter) ([]models.ResellerProfile, int64, error)
	ListDomains(filter DomainListFilter) ([]models.ResellerDomain, int64, error)
}

// SiteConfigDirectory 是管理端站点配置列表与只读查询端口。
type SiteConfigDirectory interface {
	ListSiteConfigs(filter SiteConfigListFilter) ([]models.ResellerSiteConfig, int64, error)
	GetSiteConfigByResellerID(resellerID uint) (*models.ResellerSiteConfig, error)
	GetProfileByID(id uint) (*models.ResellerProfile, error)
}
