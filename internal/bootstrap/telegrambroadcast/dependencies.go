package telegrambroadcast

import (
	"context"
	"strings"

	"github.com/dujiao-next/internal/modules/channelclient"
	broadcastapp "github.com/dujiao-next/internal/modules/telegram/broadcast/application"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/repository"
)

type UserDirectory struct {
	repository repository.UserOAuthIdentityRepository
}

func NewUserDirectory(repository repository.UserOAuthIdentityRepository) UserDirectory {
	return UserDirectory{repository: repository}
}

func (directory UserDirectory) ListTelegramUsers(query broadcastapp.UserQuery) ([]broadcastapp.UserItem, int64, error) {
	items, total, err := directory.repository.ListTelegramUsers(repository.TelegramUserListFilter{
		Page: query.Page, PageSize: query.PageSize, UserIDs: query.UserIDs,
		Keyword: query.Keyword, DisplayName: query.DisplayName,
		TelegramUsername: query.TelegramUsername, TelegramUserID: query.TelegramUserID,
		CreatedFrom: query.CreatedFrom, CreatedTo: query.CreatedTo,
	})
	if err != nil {
		return nil, 0, err
	}
	result := make([]broadcastapp.UserItem, 0, len(items))
	for _, item := range items {
		result = append(result, broadcastapp.UserItem{
			UserID: item.UserID, DisplayName: item.DisplayName, UserEmail: item.UserEmail,
			TelegramUsername: item.TelegramUsername, TelegramUserID: item.TelegramUserID,
			BoundAt: item.BoundAt, UserCreatedAt: item.UserCreatedAt,
		})
	}
	return result, total, nil
}

type BotTokenResolver struct {
	repository repository.ChannelClientRepository
	service    *channelclient.Service
}

func NewBotTokenResolver(repository repository.ChannelClientRepository, service *channelclient.Service) BotTokenResolver {
	return BotTokenResolver{repository: repository, service: service}
}

func (resolver BotTokenResolver) ResolveActiveBotToken() (string, error) {
	client, err := resolver.repository.FindActiveByChannelType("telegram_bot")
	if err != nil {
		return "", err
	}
	if client == nil || resolver.service == nil {
		return "", broadcastapp.ErrTokenUnavailable
	}
	token, err := resolver.service.DecryptBotToken(client)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(token) == "" {
		return "", broadcastapp.ErrTokenUnavailable
	}
	return strings.TrimSpace(token), nil
}

type Dispatcher struct {
	queue *queue.Client
}

func NewDispatcher(queueClient *queue.Client) Dispatcher {
	return Dispatcher{queue: queueClient}
}

func (dispatcher Dispatcher) DispatchBroadcast(_ context.Context, broadcastID uint) (bool, error) {
	if dispatcher.queue == nil || !dispatcher.queue.Enabled() {
		return false, nil
	}
	if err := dispatcher.queue.EnqueueTelegramBroadcast(queue.TelegramBroadcastPayload{BroadcastID: broadcastID}); err != nil {
		return false, err
	}
	return true, nil
}
