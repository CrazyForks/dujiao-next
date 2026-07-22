package captcha

import (
	"errors"

	settingsmodule "github.com/dujiao-next/internal/modules/settings"
)

var (
	ErrConfigInvalid = settingsmodule.ErrCaptchaConfigInvalid
	ErrRequired      = errors.New("captcha required")
	ErrInvalid       = errors.New("captcha invalid")
	ErrVerifyFailed  = errors.New("captcha verify failed")
)
