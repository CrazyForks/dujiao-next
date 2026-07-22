package telegramwiring

import (
	"context"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	telegramhttp "github.com/dujiao-next/internal/transport/http/telegram"
)

type telegramBroadcastAdapter struct {
	svc *service.TelegramBroadcastService
}

func (a telegramBroadcastAdapter) ListBroadcasts(input telegramhttp.BroadcastListInput) ([]models.TelegramBroadcast, int64, error) {
	return a.svc.ListBroadcasts(service.TelegramBroadcastListInput{
		Page:     input.Page,
		PageSize: input.PageSize,
	})
}

func (a telegramBroadcastAdapter) GetBroadcast(id uint) (*models.TelegramBroadcast, error) {
	return a.svc.GetBroadcast(id)
}

func (a telegramBroadcastAdapter) CreateBroadcast(ctx context.Context, input telegramhttp.BroadcastCreateInput) (*models.TelegramBroadcast, error) {
	return a.svc.CreateBroadcast(ctx, service.TelegramBroadcastCreateInput{
		Title:          input.Title,
		RecipientType:  input.RecipientType,
		UserIDs:        input.UserIDs,
		Filters:        input.Filters,
		AttachmentURL:  input.AttachmentURL,
		AttachmentName: input.AttachmentName,
		MessageHTML:    input.MessageHTML,
	})
}

func (a telegramBroadcastAdapter) ListTelegramUsers(input telegramhttp.BroadcastUserQuery) ([]telegramhttp.BroadcastUserItem, int64, error) {
	items, total, err := a.svc.ListTelegramUsers(service.TelegramBroadcastUserQuery{
		Page:             input.Page,
		PageSize:         input.PageSize,
		Keyword:          input.Keyword,
		DisplayName:      input.DisplayName,
		TelegramUsername: input.TelegramUsername,
		TelegramUserID:   input.TelegramUserID,
		CreatedFrom:      input.CreatedFrom,
		CreatedTo:        input.CreatedTo,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]telegramhttp.BroadcastUserItem, 0, len(items))
	for _, item := range items {
		out = append(out, telegramhttp.BroadcastUserItem{
			UserID:           item.UserID,
			DisplayName:      item.DisplayName,
			UserEmail:        item.UserEmail,
			TelegramUsername: item.TelegramUsername,
			TelegramUserID:   item.TelegramUserID,
			BoundAt:          item.BoundAt,
			UserCreatedAt:    item.UserCreatedAt,
		})
	}
	return out, total, nil
}
