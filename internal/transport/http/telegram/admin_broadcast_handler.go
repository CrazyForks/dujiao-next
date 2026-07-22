package telegramhttp

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/telegram"
	"github.com/dujiao-next/internal/shared/jsonmap"

	"github.com/gin-gonic/gin"
)

type BroadcastListInput struct {
	Page     int
	PageSize int
}

type BroadcastUserQuery struct {
	Page             int
	PageSize         int
	Keyword          string
	DisplayName      string
	TelegramUsername string
	TelegramUserID   string
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
}

type BroadcastCreateInput struct {
	Title          string
	RecipientType  string
	UserIDs        []uint
	Filters        jsonmap.JSON
	AttachmentURL  string
	AttachmentName string
	MessageHTML    string
}

type BroadcastUserItem struct {
	UserID           uint      `json:"user_id"`
	DisplayName      string    `json:"display_name"`
	UserEmail        string    `json:"user_email"`
	TelegramUsername string    `json:"telegram_username"`
	TelegramUserID   string    `json:"telegram_user_id"`
	BoundAt          time.Time `json:"bound_at"`
	UserCreatedAt    time.Time `json:"user_created_at"`
}

// BroadcastAdminService 是后台 Telegram 群发端口。
type BroadcastAdminService interface {
	ListBroadcasts(input BroadcastListInput) ([]models.TelegramBroadcast, int64, error)
	GetBroadcast(id uint) (*models.TelegramBroadcast, error)
	CreateBroadcast(ctx context.Context, input BroadcastCreateInput) (*models.TelegramBroadcast, error)
	ListTelegramUsers(input BroadcastUserQuery) ([]BroadcastUserItem, int64, error)
}

type createBroadcastRequest struct {
	Title          string       `json:"title" binding:"required"`
	RecipientType  string       `json:"recipient_type" binding:"required"`
	UserIDs        []uint       `json:"user_ids"`
	Filters        jsonmap.JSON `json:"filters"`
	AttachmentURL  string       `json:"attachment_url"`
	AttachmentName string       `json:"attachment_name"`
	MessageHTML    string       `json:"message_html" binding:"required"`
}

// AdminBroadcastHandler 处理后台 Telegram 群发请求。
type AdminBroadcastHandler struct {
	broadcasts BroadcastAdminService
}

func NewAdminBroadcastHandler(broadcasts BroadcastAdminService) *AdminBroadcastHandler {
	if broadcasts == nil {
		panic("telegram broadcast handler: broadcasts is nil")
	}
	return &AdminBroadcastHandler{broadcasts: broadcasts}
}

// ListTelegramBroadcasts 获取 Telegram 群发列表。
func (h *AdminBroadcastHandler) ListTelegramBroadcasts(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)

	items, total, err := h.broadcasts.ListBroadcasts(BroadcastListInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.bad_request", err)
		return
	}
	response.SuccessWithPage(c, items, response.BuildPagination(page, pageSize, total))
}

// CreateTelegramBroadcast 创建 Telegram 群发任务。
func (h *AdminBroadcastHandler) CreateTelegramBroadcast(c *gin.Context) {
	var req createBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	result, err := h.broadcasts.CreateBroadcast(c.Request.Context(), BroadcastCreateInput(req))
	if err != nil {
		if errors.Is(err, telegram.ErrBroadcastInvalid) ||
			errors.Is(err, telegram.ErrBroadcastNoRecipients) ||
			errors.Is(err, telegram.ErrBotTokenUnavailable) {
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.bad_request", err)
		return
	}
	response.Success(c, result)
}

// GetTelegramBroadcast 获取单条 Telegram 群发详情。
func (h *AdminBroadcastHandler) GetTelegramBroadcast(c *gin.Context) {
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", errors.New("invalid broadcast id"))
		return
	}
	broadcast, err := h.broadcasts.GetBroadcast(id)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.bad_request", err)
		return
	}
	if broadcast == nil {
		shared.RespondError(c, response.CodeNotFound, "error.not_found", telegram.ErrBroadcastNotFound)
		return
	}
	response.Success(c, broadcast)
}

// ListTelegramBroadcastUsers 获取 Telegram 广播可选用户。
func (h *AdminBroadcastHandler) ListTelegramBroadcastUsers(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)

	createdFrom, createdTo, err := shared.ParseQueryTimeRange(c, "created_from", "created_to")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}

	items, total, err := h.broadcasts.ListTelegramUsers(BroadcastUserQuery{
		Page:             page,
		PageSize:         pageSize,
		Keyword:          strings.TrimSpace(c.Query("keyword")),
		DisplayName:      strings.TrimSpace(c.Query("display_name")),
		TelegramUsername: strings.TrimSpace(c.Query("telegram_username")),
		TelegramUserID:   strings.TrimSpace(c.Query("telegram_user_id")),
		CreatedFrom:      createdFrom,
		CreatedTo:        createdTo,
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.bad_request", err)
		return
	}
	response.SuccessWithPage(c, items, response.BuildPagination(page, pageSize, total))
}
