package captchawiring

import (
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/modules/captcha"
)

// Verifier adapts the captcha module to HTTP transport contracts.
type Verifier struct {
	service *captcha.Service
}

func NewVerifier(service *captcha.Service) Verifier {
	return Verifier{service: service}
}

func (v Verifier) Verify(scene string, payload shared.CaptchaPayloadRequest, clientIP string) error {
	if v.service == nil {
		return nil
	}
	return v.service.Verify(scene, payload.ToCaptchaPayload(), clientIP)
}
