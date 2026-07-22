package router

import (
	"github.com/dujiao-next/internal/provider"
	affiliatetransport "github.com/dujiao-next/internal/transport/http/affiliate"
	channeltransport "github.com/dujiao-next/internal/transport/http/channel"
	giftcardtransport "github.com/dujiao-next/internal/transport/http/giftcard"
	memberleveltransport "github.com/dujiao-next/internal/transport/http/memberlevel"
	telegramtransport "github.com/dujiao-next/internal/transport/http/telegram"
	wallettransport "github.com/dujiao-next/internal/transport/http/wallet"

	"github.com/gin-gonic/gin"
)

func registerChannelRoutes(
	apiV1 *gin.RouterGroup,
	c *provider.Container,
	channelHandler *channeltransport.Handler,
	channelMemberLevelHandler *memberleveltransport.ChannelHandler,
	channelGiftCardHandler *giftcardtransport.ChannelHandler,
	channelAffiliateHandler *affiliatetransport.ChannelHandler,
	channelTelegramBotHandler *telegramtransport.ChannelBotHandler,
	channelWalletHandler *wallettransport.ChannelHandler,
) {
	// 渠道 API（Telegram Bot 等外部服务调用）
	channelAPI := apiV1.Group("/channel")
	channelAPI.Use(ChannelAPIAuthMiddleware(c))
	{
		telegramtransport.RegisterChannelBotRoutes(channelAPI, channelTelegramBotHandler)
		channeltransport.RegisterRoutes(channelAPI, channelHandler)
		affiliatetransport.RegisterChannelRoutes(channelAPI, channelAffiliateHandler)

		// Catalog 端点（商品浏览）
		memberleveltransport.RegisterChannelRoutes(channelAPI, channelMemberLevelHandler)

		// Order / Payment 端点（购买流程）

		// Wallet 端点（钱包）
		wallettransport.RegisterChannelRoutes(channelAPI, channelWalletHandler)
		giftcardtransport.RegisterChannelRoutes(channelAPI, channelGiftCardHandler)
	}
}
