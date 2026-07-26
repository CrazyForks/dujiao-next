package systemhttp

import "github.com/gin-gonic/gin"

// RegisterAdminRoutes 注册后台系统信息路由。
func RegisterAdminRoutes(admin gin.IRoutes, handler *AdminHandler) {
	admin.GET("/system/version/check", handler.CheckSystemUpdate)
}
