package router

import (
	"github.com/dujiao-next/internal/provider"
	upstreamtransport "github.com/dujiao-next/internal/transport/http/upstream"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func registerUpstreamRoutes(
	apiV1 *gin.RouterGroup,
	c *provider.Container,
	upstreamHandler *upstreamtransport.Handler,
	redisClient *redis.Client,
	upstreamAPIRule RateLimitRule,
) {
	// 上游 API（本站作为 B 站点，暴露给下游 A 调用）
	upstreamAPI := apiV1.Group("/upstream")
	upstreamAPI.Use(RateLimitMiddleware(redisClient, upstreamAPIRule, KeyByUpstreamApiKey))
	upstreamAPI.Use(UpstreamAPIAuthMiddleware(c.ApiCredentialRepo))
	upstreamtransport.RegisterAuthenticatedRoutes(upstreamAPI, upstreamHandler)

	// 上游回调接收（本站作为 A 站点，接收 B 的回调）
	apiV1.POST("/upstream/callback", upstreamHandler.HandleCallback)
}
