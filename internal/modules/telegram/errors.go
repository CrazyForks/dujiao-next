package telegram

import (
	"errors"

	"github.com/dujiao-next/internal/modules/notification"
)

var (
	ErrBroadcastInvalid      = errors.New("telegram broadcast invalid")
	ErrBroadcastNotFound     = errors.New("telegram broadcast not found")
	ErrBroadcastNoRecipients = errors.New("telegram broadcast no recipients")
	ErrBotTokenUnavailable   = errors.New("telegram bot token unavailable")
	ErrNotifyConfigInvalid   = notification.ErrConfigInvalid
	ErrNotifySendFailed      = notification.ErrSendFailed
)
