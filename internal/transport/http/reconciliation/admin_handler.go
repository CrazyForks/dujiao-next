package reconciliationhttp

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/reconciliation"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

type Service interface {
	CreateAndEnqueue(input reconciliation.RunInput) (*models.ReconciliationJob, error)
	ListJobs(filter reconciliation.JobListFilter) ([]models.ReconciliationJob, int64, error)
	GetJob(id uint) (*models.ReconciliationJob, error)
	GetJobItems(jobID uint, page, pageSize int) ([]models.ReconciliationItem, int64, error)
	ResolveItem(itemID, adminID uint, remark string) error
}

type AdminHandler struct {
	service Service
}

func NewAdminHandler(service Service) *AdminHandler {
	if service == nil {
		panic("reconciliation admin handler: required dependency is nil")
	}
	return &AdminHandler{service: service}
}

func (h *AdminHandler) Run(c *gin.Context) {
	var input reconciliation.RunInput
	if err := c.ShouldBindJSON(&input); err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	job, err := h.service.CreateAndEnqueue(input)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.reconciliation_create_failed", err)
		return
	}
	response.Success(c, job)
}

func (h *AdminHandler) ListJobs(c *gin.Context) {
	page, pageSize := ginutil.ParsePagination(c)
	filter := reconciliation.JobListFilter{Page: page, PageSize: pageSize}
	if connectionID := strings.TrimSpace(c.Query("connection_id")); connectionID != "" {
		if id, err := ginutil.ParseQueryUint(connectionID, false); err == nil {
			filter.ConnectionID = id
		}
	}
	filter.Status = strings.TrimSpace(c.Query("status"))
	filter.Type = strings.TrimSpace(c.Query("type"))
	jobs, total, err := h.service.ListJobs(filter)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.reconciliation_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, jobs, response.BuildPagination(page, pageSize, total))
}

func (h *AdminHandler) GetJob(c *gin.Context) {
	id, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	job, err := h.service.GetJob(id)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.reconciliation_fetch_failed", err)
		return
	}
	itemPage, itemPageSize := ginutil.ParsePaginationWithKeys(c, "items_page", "items_page_size", 20)
	items, total, err := h.service.GetJobItems(id, itemPage, itemPageSize)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.reconciliation_fetch_failed", err)
		return
	}
	response.Success(c, gin.H{"job": job, "items": items, "items_total": total})
}

func (h *AdminHandler) ResolveItem(c *gin.Context) {
	id, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	var input struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	adminID, ok := ginutil.GetAdminID(c)
	if !ok {
		ginutil.RespondError(c, response.CodeUnauthorized, "error.unauthorized", nil)
		return
	}
	if err := h.service.ResolveItem(id, adminID, input.Remark); err != nil {
		if errors.Is(err, reconciliation.ErrItemNotFound) {
			ginutil.RespondError(c, response.CodeNotFound, "error.reconciliation_item_not_found", nil)
			return
		}
		ginutil.RespondError(c, response.CodeInternal, "error.reconciliation_resolve_failed", err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}
