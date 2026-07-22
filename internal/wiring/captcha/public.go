package captchawiring

import (
	"github.com/dujiao-next/internal/provider"
	captchahttp "github.com/dujiao-next/internal/transport/http/captcha"
)

func NewPublicHandler(c *provider.Container) *captchahttp.PublicHandler {
	return captchahttp.NewPublicHandler(c.CaptchaService)
}
