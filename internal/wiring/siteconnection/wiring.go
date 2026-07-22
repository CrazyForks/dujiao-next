package siteconnectionwiring

import (
	"github.com/dujiao-next/internal/service"
)

type siteConnectionMarkupAdapter struct {
	svc *service.ProductMappingService
}

func (a siteConnectionMarkupAdapter) ReapplyMarkup(connectionID uint) (int, error) {
	return a.svc.ReapplyMarkup(connectionID)
}
