package resellerwiring

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	resellerhttp "github.com/dujiao-next/internal/transport/http/reseller"
)

type resellerAdminDirectoryAdapter struct {
	repo repository.ResellerRepository
}

func (a resellerAdminDirectoryAdapter) ListProfiles(filter resellerhttp.ProfileListFilter) ([]models.ResellerProfile, int64, error) {
	return a.repo.ListProfiles(repository.ResellerProfileListFilter{
		Page:             filter.Page,
		PageSize:         filter.PageSize,
		UserID:           filter.UserID,
		Status:           filter.Status,
		SettlementStatus: filter.SettlementStatus,
		Keyword:          filter.Keyword,
		CreatedFrom:      filter.CreatedFrom,
		CreatedTo:        filter.CreatedTo,
	})
}

func (a resellerAdminDirectoryAdapter) ListDomains(filter resellerhttp.DomainListFilter) ([]models.ResellerDomain, int64, error) {
	return a.repo.ListDomains(repository.ResellerDomainListFilter{
		Page:               filter.Page,
		PageSize:           filter.PageSize,
		ResellerID:         filter.ResellerID,
		UserID:             filter.UserID,
		Domain:             filter.Domain,
		Type:               filter.Type,
		Status:             filter.Status,
		VerificationStatus: filter.VerificationStatus,
		Keyword:            filter.Keyword,
		CreatedFrom:        filter.CreatedFrom,
		CreatedTo:          filter.CreatedTo,
	})
}

func (a resellerAdminDirectoryAdapter) ListSiteConfigs(filter resellerhttp.SiteConfigListFilter) ([]models.ResellerSiteConfig, int64, error) {
	return a.repo.ListSiteConfigs(repository.ResellerSiteConfigListFilter{
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		ResellerID:  filter.ResellerID,
		Keyword:     filter.Keyword,
		CreatedFrom: filter.CreatedFrom,
		CreatedTo:   filter.CreatedTo,
	})
}

func (a resellerAdminDirectoryAdapter) GetSiteConfigByResellerID(resellerID uint) (*models.ResellerSiteConfig, error) {
	return a.repo.GetSiteConfigByResellerID(resellerID)
}

func (a resellerAdminDirectoryAdapter) GetProfileByID(id uint) (*models.ResellerProfile, error) {
	return a.repo.GetProfileByID(id)
}

func (a resellerAdminDirectoryAdapter) ListDomainsByResellerID(resellerID uint) ([]models.ResellerDomain, error) {
	return a.repo.ListDomainsByResellerID(resellerID)
}
