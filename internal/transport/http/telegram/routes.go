package telegramhttp

import "github.com/gin-gonic/gin"

// RegisterAdminBroadcastRoutes 注册后台 Telegram 群发路由。
func RegisterAdminBroadcastRoutes(admin gin.IRoutes, handler *AdminBroadcastHandler) {
	admin.GET("/telegram-bot/broadcasts", handler.ListTelegramBroadcasts)
	admin.GET("/telegram-bot/broadcasts/:id", handler.GetTelegramBroadcast)
	admin.POST("/telegram-bot/broadcasts", handler.CreateTelegramBroadcast)
	admin.GET("/telegram-bot/users", handler.ListTelegramBroadcastUsers)
}

// RegisterChannelBotRoutes 注册渠道 Telegram Bot 配置与心跳路由。
func RegisterChannelBotRoutes(channel gin.IRoutes, handler *ChannelBotHandler) {
	channel.GET("/telegram/config", handler.GetBotConfig)
	channel.POST("/telegram/heartbeat", handler.ReportHeartbeat)
}
