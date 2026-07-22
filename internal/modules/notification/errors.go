package notification

import (
	"errors"

	settingsmessaging "github.com/dujiao-next/internal/modules/settings/schema/messaging"
)

var (
	ErrConfigInvalid = settingsmessaging.ErrNotificationConfigInvalid
	ErrSendFailed    = errors.New("notification send failed")
	ErrEventInvalid  = errors.New("notification event invalid")
)
