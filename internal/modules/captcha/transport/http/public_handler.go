package captchahttp

import (
	"errors"

	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

// ImageChallengeGenerator 生成图片验证码挑战。
type ImageChallengeGenerator interface {
	GenerateImageChallenge() (*captcha.ImageChallenge, error)
}

// PublicHandler 处理公开验证码请求。
type PublicHandler struct {
	generator ImageChallengeGenerator
}

func NewPublicHandler(generator ImageChallengeGenerator) *PublicHandler {
	return &PublicHandler{generator: generator}
}

// GetImageCaptcha 获取图片验证码挑战。
func (h *PublicHandler) GetImageCaptcha(c *gin.Context) {
	if h == nil || h.generator == nil {
		ginutil.RespondError(c, response.CodeInternal, "error.captcha_unavailable", captcha.ErrConfigInvalid)
		return
	}

	challenge, err := h.generator.GenerateImageChallenge()
	if err != nil {
		switch {
		case errors.Is(err, captcha.ErrConfigInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.captcha_unavailable", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.captcha_generate_failed", err)
		}
		return
	}

	response.Success(c, gin.H{
		"captcha_id":   challenge.CaptchaID,
		"image_base64": challenge.ImageBase64,
	})
}
