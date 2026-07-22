package siteconnectionwiring

type siteConnectionMarkupAdapter struct {
	svc interface {
		ReapplyMarkup(connectionID uint) (int, error)
	}
}

func (a siteConnectionMarkupAdapter) ReapplyMarkup(connectionID uint) (int, error) {
	return a.svc.ReapplyMarkup(connectionID)
}
