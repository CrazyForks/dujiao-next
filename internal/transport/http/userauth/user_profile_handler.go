package userauthhttp

import (
	"errors"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

var (
	ErrProfileEmpty = errors.New("user profile empty")
	ErrUserNotFound = errors.New("user not found")
)

// UserProfileService 是用户资料端点所需的最小端口。
type UserProfileService interface {
	GetUserByID(id uint) (*models.User, error)
	ResolveEmailChangeMode(user *models.User) (string, error)
	ResolvePasswordChangeMode(user *models.User) (string, error)
	UpdateProfile(userID uint, nickname, locale *string) (*models.User, error)
}

// UserProfileHandler 处理当前用户资料 HTTP 请求。
type UserProfileHandler struct {
	service UserProfileService
}

func NewUserProfileHandler(service UserProfileService) *UserProfileHandler {
	if service == nil {
		panic("user profile handler: service is nil")
	}
	return &UserProfileHandler{service: service}
}

// UserProfileUpdateRequest 更新资料请求。
type UserProfileUpdateRequest struct {
	Nickname *string `json:"nickname"`
	Locale   *string `json:"locale"`
}

// GetCurrentUser 获取当前用户信息。
func (h *UserProfileHandler) GetCurrentUser(c *gin.Context) {
	id, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	user, err := h.service.GetUserByID(id)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	if user == nil {
		shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		return
	}

	profile, err := h.userProfileResponse(user)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.Success(c, profile)
}

func (h *UserProfileHandler) userProfileResponse(user *models.User) (dto.UserProfileResp, error) {
	emailMode, err := h.service.ResolveEmailChangeMode(user)
	if err != nil {
		return dto.UserProfileResp{}, err
	}
	passwordMode, err := h.service.ResolvePasswordChangeMode(user)
	if err != nil {
		return dto.UserProfileResp{}, err
	}
	return dto.NewUserProfileResp(user, emailMode, passwordMode), nil
}

// UpdateUserProfile 更新用户资料。
func (h *UserProfileHandler) UpdateUserProfile(c *gin.Context) {
	id, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	var req UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	user, err := h.service.UpdateProfile(id, req.Nickname, req.Locale)
	if err != nil {
		switch {
		case errors.Is(err, ErrProfileEmpty):
			shared.RespondError(c, response.CodeBadRequest, "error.profile_empty", nil)
		case errors.Is(err, ErrUserNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.user_update_failed", err)
		}
		return
	}

	profile, err := h.userProfileResponse(user)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_update_failed", err)
		return
	}
	response.Success(c, profile)
}
