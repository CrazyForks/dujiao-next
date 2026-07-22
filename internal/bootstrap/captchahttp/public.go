package captchabootstrap

import (
	captchahttp "github.com/dujiao-next/internal/modules/captcha/transport/http"
	"github.com/dujiao-next/internal/provider"
)

func NewPublicHandler(c *provider.Container) *captchahttp.PublicHandler {
	return captchahttp.NewPublicHandler(c.CaptchaService)
}
