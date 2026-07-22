package reselleradmin_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	"github.com/dujiao-next/internal/shared/jsonmap"
	settingstransport "github.com/dujiao-next/internal/transport/http/settings"

	"github.com/gin-gonic/gin"
)

type adminSettingRepository struct {
	store map[string]jsonmap.JSON
}

func newAdminSettingRepository() *adminSettingRepository {
	return &adminSettingRepository{store: make(map[string]jsonmap.JSON)}
}

func (repository *adminSettingRepository) GetByKey(key string) (*models.Setting, error) {
	value, exists := repository.store[key]
	if !exists {
		return nil, nil
	}
	return &models.Setting{Key: key, ValueJSON: value}, nil
}

func (repository *adminSettingRepository) Upsert(key string, value jsonmap.JSON) (*models.Setting, error) {
	repository.store[key] = value
	return &models.Setting{Key: key, ValueJSON: value}, nil
}

func TestUpdateSettingsInvalidatesCallbackRoutesFromRegistryEffect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := newAdminSettingRepository()
	settingService := service.NewSettingService(repository)
	settingService.InvalidateCallbackRoutesCache()
	t.Cleanup(settingService.InvalidateCallbackRoutesCache)

	if _, err := settingService.Update(constants.SettingKeyCallbackRoutesConfig, map[string]interface{}{
		constants.SettingFieldPaymentCallback: "/api/old/callback",
	}); err != nil {
		t.Fatalf("seed callback routes: %v", err)
	}
	if cached := settingService.GetCallbackRoutesCached(); cached == nil || cached.PaymentCallback != "/api/old/callback" {
		t.Fatalf("seed callback route cache mismatch: %#v", cached)
	}

	handler := settingstransport.NewAdminHandler(settingService)
	router := gin.New()
	settingstransport.RegisterAdminRoutes(router, handler)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewBufferString(`{
		"key":"callback_routes_config",
		"value":{"payment_callback":"/api/new/callback"}
	}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("update settings HTTP status want 200 got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if cached := settingService.GetCallbackRoutesCached(); cached == nil || cached.PaymentCallback != "/api/new/callback" {
		t.Fatalf("callback cache was not refreshed after update: %#v", cached)
	}
}
