package compliancewiring

import (
	"github.com/dujiao-next/internal/provider"
	compliancetransport "github.com/dujiao-next/internal/transport/http/compliance"
)

func NewAdminHandler(c *provider.Container) *compliancetransport.AdminHandler {
	return compliancetransport.NewAdminHandler(complianceAdminAdapter{svc: c.ComplianceService})
}
