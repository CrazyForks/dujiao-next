package auditloghttp

import (
	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

type UserLoginHistoryReader interface {
	ListByUser(userID uint, page, pageSize int) ([]models.UserLoginLog, int64, error)
}

type UserHandler struct {
	userLoginLogs UserLoginHistoryReader
}

func NewUserHandler(userLoginLogs UserLoginHistoryReader) *UserHandler {
	return &UserHandler{userLoginLogs: userLoginLogs}
}

// GetMyLoginLogs 获取当前用户登录日志
func (h *UserHandler) GetMyLoginLogs(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	page, pageSize := shared.ParsePagination(c)

	logs, total, err := h.userLoginLogs.ListByUser(uid, page, pageSize)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_login_log_fetch_failed", err)
		return
	}

	pagination := response.BuildPagination(page, pageSize, total)
	response.SuccessWithPage(c, dto.NewLoginLogRespList(logs), pagination)
}
