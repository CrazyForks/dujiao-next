package captchahttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/modules/captcha"

	"github.com/gin-gonic/gin"
)

type fakeGenerator struct {
	challenge *captcha.ImageChallenge
	err       error
}

func (f fakeGenerator) GenerateImageChallenge() (*captcha.ImageChallenge, error) {
	return f.challenge, f.err
}

func TestGetImageCaptchaSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(fakeGenerator{challenge: &captcha.ImageChallenge{
		CaptchaID:   "id-1",
		ImageBase64: "data:image/png;base64,abc",
	}})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeOK {
		t.Fatalf("status want OK got %d body=%s", got.StatusCode, w.Body.String())
	}
	data, ok := got.Data.(map[string]interface{})
	if !ok || data["captcha_id"] != "id-1" {
		t.Fatalf("unexpected data %#v", got.Data)
	}
}

func TestGetImageCaptchaUnavailableWhenConfigInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(fakeGenerator{err: captcha.ErrConfigInvalid})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeBadRequest {
		t.Fatalf("status want 400 got %d", got.StatusCode)
	}
}

func TestGetImageCaptchaUnavailableWithoutGenerator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeInternal {
		t.Fatalf("status want internal got %d", got.StatusCode)
	}
}

func TestGetImageCaptchaMapsUnknownErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPublicHandler(fakeGenerator{err: errors.New("boom")})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/captcha/image", nil)

	handler.GetImageCaptcha(c)

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.StatusCode != response.CodeInternal {
		t.Fatalf("status want internal got %d", got.StatusCode)
	}
}
