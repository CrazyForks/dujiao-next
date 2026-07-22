package resellerhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"

	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

type profileDetailDirectoryStub struct {
	profile *models.ResellerProfile
}

func (s profileDetailDirectoryStub) GetProfileByID(id uint) (*models.ResellerProfile, error) {
	return s.profile, nil
}
func (s profileDetailDirectoryStub) ListDomainsByResellerID(resellerID uint) ([]models.ResellerDomain, error) {
	return nil, nil
}
func (s profileDetailDirectoryStub) GetSiteConfigByResellerID(resellerID uint) (*models.ResellerSiteConfig, error) {
	return nil, nil
}

type orderAdminListerStub struct {
	err error
}

func (s orderAdminListerStub) ListAdminOrders(resellerID uint, input resellermodule.OrderListInput) ([]resellermodule.OrderListItem, int64, error) {
	return nil, 0, s.err
}

func TestAdminProfileDetailHandlerMapsOrderNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAdminProfileDetailHandler(
		profileDetailDirectoryStub{profile: &models.ResellerProfile{ID: 1}},
		nil,
		nil,
		orderAdminListerStub{err: productcontract.ErrNotFound},
	)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/resellers/profiles/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.GetProfileDetail(c)

	var resp struct {
		StatusCode int `json:"status_code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.StatusCode != response.CodeNotFound {
		t.Fatalf("expected not found, body=%s", recorder.Body.String())
	}
}
