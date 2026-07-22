package sitemapwiring

import "github.com/dujiao-next/internal/service"

type sitemapSiteBrandAdapter struct {
	settings *service.SettingService
}

func (a sitemapSiteBrandAdapter) GetSiteURL() (string, error) {
	if a.settings == nil {
		return "", nil
	}
	brand, err := a.settings.GetSiteBrand()
	if err != nil {
		return "", err
	}
	return brand.SiteURL, nil
}
