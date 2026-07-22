package telegramwiring

import "github.com/dujiao-next/internal/modules/channelclient"

// telegramChannelBotTokenAdapter 将 ChannelClientService 适配为 bot token 解密端口。
type telegramChannelBotTokenAdapter struct {
	svc *channelclient.Service
}

func (a telegramChannelBotTokenAdapter) DecryptBotTokenByClientID(clientID uint) (string, error) {
	client, err := a.svc.GetChannelClient(clientID)
	if err != nil {
		return "", err
	}
	if client == nil {
		return "", channelclient.ErrNotFound
	}
	return a.svc.DecryptBotToken(client)
}
