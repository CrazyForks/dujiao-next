package adminauthhttp

import (
	"errors"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

// UserTOTPService 是管理员重置用户 2FA 端口。
type UserTOTPService interface {
	AdminResetUser2FA(operatorID, userID uint) (*models.User, error)
}

// AdminUser2FAHandler 处理管理员对用户 2FA 的管理请求。
type AdminUser2FAHandler struct {
	userTOTP UserTOTPService
}

func NewAdminUser2FAHandler(userTOTP UserTOTPService) *AdminUser2FAHandler {
	if userTOTP == nil {
		panic("admin user 2fa handler: userTOTP is nil")
	}
	return &AdminUser2FAHandler{userTOTP: userTOTP}
}

// ResetUser2FA 管理员重置目标用户 2FA。
// DELETE /admin/users/:id/2fa
func (h *AdminUser2FAHandler) ResetUser2FA(c *gin.Context) {
	operatorID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	userID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.user_id_invalid", nil)
		return
	}
	target, err := h.userTOTP.AdminResetUser2FA(operatorID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		case errors.Is(err, ErrTOTPNotEnabled):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_not_enabled", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		}
		return
	}

	requestID, _ := c.Get("request_id")
	rid, _ := requestID.(string)
	logger.Warnw("admin_reset_user_2fa",
		"operator_admin_id", operatorID,
		"target_user_id", target.ID,
		"target_email", target.Email,
		"client_ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
		"request_id", rid,
	)

	response.Success(c, nil)
}
