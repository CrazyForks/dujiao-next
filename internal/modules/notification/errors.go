package notification

import (
	"errors"

	settingsmodule "github.com/dujiao-next/internal/modules/settings"
)

var (
	ErrConfigInvalid = settingsmodule.ErrNotificationConfigInvalid
	ErrSendFailed    = errors.New("notification send failed")
	ErrEventInvalid  = errors.New("notification event invalid")
)
