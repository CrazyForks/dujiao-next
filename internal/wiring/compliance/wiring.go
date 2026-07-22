package compliancewiring

import (
	"github.com/dujiao-next/internal/modules/compliance"
)

// complianceAdminAdapter 将 legacy ComplianceService 适配为 transport 端口。
type complianceAdminAdapter struct {
	svc *compliance.Service
}

func (a complianceAdminAdapter) Status() (*compliance.Status, error) {
	return a.svc.Status()
}

func (a complianceAdminAdapter) Acknowledge(req compliance.AcknowledgeRequest) error {
	return a.svc.Acknowledge(req)
}
