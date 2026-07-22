package sitemapwiring

import settingsapp "github.com/dujiao-next/internal/modules/settings/application"

type sitemapSiteBrandAdapter struct {
	settings *settingsapp.Service
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
