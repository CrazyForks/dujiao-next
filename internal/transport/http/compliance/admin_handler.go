package compliancehttp

import (
	"errors"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/modules/compliance"

	"github.com/gin-gonic/gin"
)

// AdminService 是后台合规声明端口。
type AdminService interface {
	Status() (*compliance.Status, error)
	Acknowledge(req compliance.AcknowledgeRequest) error
}

type acknowledgeRequest struct {
	Segment1 string `json:"segment1" binding:"required"`
	Segment2 string `json:"segment2" binding:"required"`
	Segment3 string `json:"segment3" binding:"required"`
}

// AdminHandler 处理后台合规声明确认请求。
type AdminHandler struct {
	svc AdminService
}

func NewAdminHandler(svc AdminService) *AdminHandler {
	if svc == nil {
		panic("compliance admin handler: service is nil")
	}
	return &AdminHandler{svc: svc}
}

// GetComplianceStatus GET /admin/compliance/status
func (h *AdminHandler) GetComplianceStatus(c *gin.Context) {
	status, err := h.svc.Status()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, status)
}

// AcknowledgeCompliance POST /admin/compliance/acknowledge —— 仅超管
func (h *AdminHandler) AcknowledgeCompliance(c *gin.Context) {
	if !shared.IsSuperAdmin(c) {
		shared.RespondError(c, response.CodeForbidden, "compliance.error.super_admin_required", nil)
		return
	}
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	username := ""
	if v, exists := c.Get("username"); exists {
		username, _ = v.(string)
	}

	var req acknowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	err := h.svc.Acknowledge(compliance.AcknowledgeRequest{
		Segment1:  req.Segment1,
		Segment2:  req.Segment2,
		Segment3:  req.Segment3,
		AdminID:   adminID,
		Username:  username,
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, compliance.ErrTextMismatch):
			shared.RespondError(c, response.CodeBadRequest, "compliance.error.text_mismatch", nil)
			return
		case errors.Is(err, compliance.ErrAlreadyAcknowledged):
			response.Success(c, gin.H{"already_acknowledged": true})
			return
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal", err)
			return
		}
	}
	response.Success(c, gin.H{"already_acknowledged": false})
}
