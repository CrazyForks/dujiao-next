package systemhttp

import (
	"errors"

	ginutil "github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/selfupdate"

	"github.com/gin-gonic/gin"
)

// GetUpdateCapability 返回当前部署环境是否支持一键升级。
// 前端据此决定展示升级按钮还是手动升级指引（容器部署走 compose 命令）。
// GET /api/v1/admin/system/update/capability
func (h *AdminHandler) GetUpdateCapability(c *gin.Context) {
	response.Success(c, gin.H{
		"capability": selfupdate.Detect(),
		"state":      h.updates.Snapshot(),
	})
}

// StartUpdate 启动一键升级任务。任务在后台执行，进度由 GetUpdateStatus 轮询。
// POST /api/v1/admin/system/update/start
func (h *AdminHandler) StartUpdate(c *gin.Context) {
	err := h.updates.Start(c.Request.Context())
	switch {
	case err == nil:
		state := h.updates.Snapshot()
		// 替换二进制是不可逆的高危操作，留下操作者与目标版本便于事后追溯
		logOperation(c, "self_update_started", "target_version", state.TargetVersion)
		response.Success(c, state)
	case errors.Is(err, selfupdate.ErrNotSupported):
		ginutil.RespondError(c, response.CodeBadRequest, "error.update_not_supported", err)
	case errors.Is(err, selfupdate.ErrUpdateInProgress):
		ginutil.RespondError(c, response.CodeBadRequest, "error.update_in_progress", err)
	case errors.Is(err, selfupdate.ErrNoUpdateAvailable):
		ginutil.RespondError(c, response.CodeBadRequest, "error.update_already_latest", err)
	default:
		ginutil.RespondError(c, response.CodeInternal, "error.update_failed", err)
	}
}

// GetUpdateStatus 轮询升级任务进度
// GET /api/v1/admin/system/update/status
func (h *AdminHandler) GetUpdateStatus(c *gin.Context) {
	response.Success(c, h.updates.Snapshot())
}

// RollbackUpdate 还原到升级前的二进制。
// 仅在新版本启动失败时有意义 —— 新版本一旦跑通并完成迁移，退回旧二进制未必兼容。
// POST /api/v1/admin/system/update/rollback
func (h *AdminHandler) RollbackUpdate(c *gin.Context) {
	err := h.updates.Rollback()
	switch {
	case err == nil:
		logOperation(c, "self_update_rolled_back")
		response.Success(c, h.updates.Snapshot())
	case errors.Is(err, selfupdate.ErrNoBackup):
		ginutil.RespondError(c, response.CodeBadRequest, "error.update_no_backup", err)
	case errors.Is(err, selfupdate.ErrUpdateInProgress):
		ginutil.RespondError(c, response.CodeBadRequest, "error.update_in_progress", err)
	default:
		ginutil.RespondError(c, response.CodeInternal, "error.update_rollback_failed", err)
	}
}

// RestartService 让进程退出，由 systemd 拉起替换后的新二进制。
// 没有守护进程时拒绝执行 —— 那种情况下退出等于停服。
// POST /api/v1/admin/system/restart
func (h *AdminHandler) RestartService(c *gin.Context) {
	if h.updates.Running() {
		ginutil.RespondError(c, response.CodeBadRequest, "error.update_in_progress", selfupdate.ErrUpdateInProgress)
		return
	}
	if !selfupdate.Detect().CanRestart {
		ginutil.RespondError(c, response.CodeBadRequest, "error.restart_not_supported", selfupdate.ErrNotSupported)
		return
	}

	logOperation(c, "self_update_restart_requested")

	// 先应答再退出：Restart 内部延迟半秒发 SIGTERM，
	// 保证这条响应能写回客户端，前端才知道该进入等待重连状态。
	response.Success(c, gin.H{"restarting": true})
	selfupdate.Restart()
}

// logOperation 记录升级类操作的操作者。这类接口会替换程序文件或重启进程，
// 出问题时需要能回答「谁在什么时候动了什么版本」。
func logOperation(c *gin.Context, action string, kv ...any) {
	adminID, _ := ginutil.GetAdminID(c)
	fields := append([]any{"admin_id", adminID, "client_ip", c.ClientIP()}, kv...)
	ginutil.RequestLog(c).Infow(action, fields...)
}
