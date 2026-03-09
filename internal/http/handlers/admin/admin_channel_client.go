package admin

import (
	"errors"
	"strconv"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

// ListChannelClients 获取渠道客户端列表（含解密 secret）
func (h *Handler) ListChannelClients(c *gin.Context) {
	clients, err := h.ChannelClientService.ListChannelClientDetails()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.channel_clients_fetch_failed", err)
		return
	}
	response.Success(c, clients)
}

type createChannelClientRequest struct {
	Name        string `json:"name" binding:"required"`
	ChannelType string `json:"channel_type" binding:"required"`
	Description string `json:"description"`
	BotToken    string `json:"bot_token"`
}

// CreateChannelClient 创建渠道客户端
func (h *Handler) CreateChannelClient(c *gin.Context) {
	var req createChannelClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	result, err := h.ChannelClientService.CreateChannelClient(req.Name, req.ChannelType, req.Description, req.BotToken)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.channel_client_create_failed", err)
		return
	}

	response.Success(c, result)
}

// GetChannelClient 获取渠道客户端详情（含解密 secret）
func (h *Handler) GetChannelClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	detail, err := h.ChannelClientService.GetChannelClientDetail(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrChannelClientNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.channel_client_fetch_failed", err)
		return
	}
	response.Success(c, detail)
}

type updateChannelClientStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// UpdateChannelClientStatus 更新渠道客户端状态
func (h *Handler) UpdateChannelClientStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req updateChannelClientStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	if err := h.ChannelClientService.UpdateChannelClientStatus(uint(id), req.Status); err != nil {
		if errors.Is(err, service.ErrChannelClientNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.channel_client_update_failed", err)
		return
	}

	response.Success(c, nil)
}

type updateChannelClientRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	BotToken    *string `json:"bot_token"` // nil = 不修改, "" = 清空, "xxx" = 设置
}

// UpdateChannelClient 更新渠道客户端信息
func (h *Handler) UpdateChannelClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req updateChannelClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	result, err := h.ChannelClientService.UpdateChannelClient(uint(id), req.Name, req.Description, req.BotToken)
	if err != nil {
		if errors.Is(err, service.ErrChannelClientNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.channel_client_update_failed", err)
		return
	}

	response.Success(c, result)
}

// ResetChannelClientSecret 重置渠道客户端 Secret
func (h *Handler) ResetChannelClientSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	result, err := h.ChannelClientService.ResetChannelClientSecret(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrChannelClientNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.channel_client_reset_secret_failed", err)
		return
	}

	response.Success(c, result)
}

// DeleteChannelClient 删除渠道客户端
func (h *Handler) DeleteChannelClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	if err := h.ChannelClientService.DeleteChannelClient(uint(id)); err != nil {
		if errors.Is(err, service.ErrChannelClientNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.channel_client_delete_failed", err)
		return
	}

	response.Success(c, nil)
}
