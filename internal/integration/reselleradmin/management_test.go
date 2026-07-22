package reselleradmin_test

import (
	"encoding/json"
	"fmt"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/auditlog"
	auditloggormstore "github.com/dujiao-next/internal/modules/auditlog/store/gormstore"
	admindomain "github.com/dujiao-next/internal/modules/identity/admin/domain"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	resellerpersistence "github.com/dujiao-next/internal/persistence/reseller"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/provider"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"
	resellerhttp "github.com/dujiao-next/internal/transport/http/reseller"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type adminResellerFixture struct {
	*provider.Container
}

func setupAdminResellerManagementHandlerTest(t *testing.T) (*adminResellerFixture, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:admin_reseller_management_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&userdomain.User{},
		&admindomain.Admin{},
		&models.AuthzAuditLog{},
		&models.Category{},
		&models.Product{},
		&models.ProductSKU{},
		&models.Order{},
		&models.OrderItem{},
		&models.ResellerProfile{},
		&models.ResellerDomain{},
		&models.ResellerSiteConfig{},
		&models.ResellerProductSetting{},
		&models.ResellerOrderSnapshot{},
		&models.ResellerLedgerEntry{},
		&models.ResellerWithdrawRequest{},
		&models.ResellerBalanceAccount{},
	); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_reseller_domains_active_domain ON reseller_domains(domain) WHERE deleted_at IS NULL").Error; err != nil {
		t.Fatalf("create domain index failed: %v", err)
	}

	resellerRepo := repository.NewResellerRepository(db)
	settingRepo := repository.NewResellerProductSettingRepository(db)
	productRepo := productgormstore.NewProductStore(db)
	auditRepo := auditloggormstore.NewAuthzStore(db)
	return &adminResellerFixture{Container: &provider.Container{
		ResellerRepo:               resellerRepo,
		ResellerProductSettingRepo: settingRepo,
		ProductRepo:                productRepo,
		ResellerManagementService: resellermodule.NewManagementService(resellerpersistence.NewManagementStore(resellerRepo), config.ResellerConfig{
			Enabled:          true,
			SelfApplyEnabled: true,
			SubdomainBase:    "shop.example.test",
			MainHosts:        []string{"main.example.test"},
		}),
		ResellerProductSettingService: resellermodule.NewProductSettingService(resellerpersistence.NewProductSettingStore(settingRepo, resellerRepo), productRepo),
		ResellerAccountingService:     service.NewResellerAccountingService(resellerRepo, service.ResellerAccountingOptions{ConfirmDays: 7}),
		ResellerOrderService:          resellermodule.NewOrderQueryService(resellerRepo),
		AuthzAuditService:             auditlog.NewAuthzService(auditRepo),
	}}, db
}

func newAdminResellerManagementContext(method, target string, body *strings.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	if body == nil {
		body = strings.NewReader("")
	}
	c.Request = httptest.NewRequest(method, target, body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("admin_id", uint(99))
	c.Set("username", "finance-admin")
	c.Set("request_id", "req-reseller-management")
	return c, recorder
}

type adminResellerTestDirectory struct {
	repo repository.ResellerRepository
}

func (a adminResellerTestDirectory) ListProfiles(filter resellerhttp.ProfileListFilter) ([]models.ResellerProfile, int64, error) {
	return a.repo.ListProfiles(repository.ResellerProfileListFilter{
		Page: filter.Page, PageSize: filter.PageSize, UserID: filter.UserID,
		Status: filter.Status, SettlementStatus: filter.SettlementStatus, Keyword: filter.Keyword,
		CreatedFrom: filter.CreatedFrom, CreatedTo: filter.CreatedTo,
	})
}

func (a adminResellerTestDirectory) ListDomains(filter resellerhttp.DomainListFilter) ([]models.ResellerDomain, int64, error) {
	return a.repo.ListDomains(repository.ResellerDomainListFilter{
		Page: filter.Page, PageSize: filter.PageSize, ResellerID: filter.ResellerID, UserID: filter.UserID,
		Domain: filter.Domain, Type: filter.Type, Status: filter.Status, VerificationStatus: filter.VerificationStatus,
		Keyword: filter.Keyword, CreatedFrom: filter.CreatedFrom, CreatedTo: filter.CreatedTo,
	})
}

func (a adminResellerTestDirectory) ListSiteConfigs(filter resellerhttp.SiteConfigListFilter) ([]models.ResellerSiteConfig, int64, error) {
	return a.repo.ListSiteConfigs(repository.ResellerSiteConfigListFilter{
		Page: filter.Page, PageSize: filter.PageSize, ResellerID: filter.ResellerID, Keyword: filter.Keyword,
		CreatedFrom: filter.CreatedFrom, CreatedTo: filter.CreatedTo,
	})
}

func (a adminResellerTestDirectory) GetSiteConfigByResellerID(resellerID uint) (*models.ResellerSiteConfig, error) {
	return a.repo.GetSiteConfigByResellerID(resellerID)
}

func (a adminResellerTestDirectory) GetProfileByID(id uint) (*models.ResellerProfile, error) {
	return a.repo.GetProfileByID(id)
}

func (a adminResellerTestDirectory) ListDomainsByResellerID(resellerID uint) ([]models.ResellerDomain, error) {
	return a.repo.ListDomainsByResellerID(resellerID)
}

func adminResellerManagementHTTP(h *adminResellerFixture) *resellerhttp.AdminManagementHandler {
	return resellerhttp.NewAdminManagementHandler(h.ResellerManagementService, adminResellerTestDirectory{repo: h.ResellerRepo}, h.AuthzAuditService)
}

func adminResellerProfileDetailHTTP(h *adminResellerFixture) *resellerhttp.AdminProfileDetailHandler {
	return resellerhttp.NewAdminProfileDetailHandler(
		adminResellerTestDirectory{repo: h.ResellerRepo},
		h.ResellerProductSettingService,
		h.ResellerAccountingService,
		h.ResellerOrderService,
	)
}

func adminResellerSiteConfigHTTP(h *adminResellerFixture) *resellerhttp.AdminSiteConfigHandler {
	return resellerhttp.NewAdminSiteConfigHandler(h.ResellerSiteConfigService, adminResellerTestDirectory{repo: h.ResellerRepo}, h.AuthzAuditService)
}

func seedAdminResellerManagementProfile(t *testing.T, db *gorm.DB, status string) models.ResellerProfile {
	t.Helper()
	user := userdomain.User{Email: fmt.Sprintf("reseller-management-%d@example.test", time.Now().UnixNano()), PasswordHash: "hash", DisplayName: "Reseller Management"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	profile := models.ResellerProfile{UserID: user.ID, Status: status, ApplyReason: "please approve", SettlementStatus: models.ResellerSettlementStatusNormal}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile failed: %v", err)
	}
	return profile
}

func TestAdminResellerManagementApproveProfileActivatesWithoutAutoDomain(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusPendingReview)
	c, recorder := newAdminResellerManagementContext(http.MethodPost, fmt.Sprintf("/admin/resellers/profiles/%d/approve", profile.ID), strings.NewReader(`{"default_markup_percent":"8.00","max_markup_percent":"40.00"}`))
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", profile.ID)}}

	adminResellerManagementHTTP(h).ApproveProfile(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var domainCount int64
	if err := db.Model(&models.ResellerDomain{}).Where("reseller_id = ?", profile.ID).Count(&domainCount).Error; err != nil {
		t.Fatalf("count domains failed: %v", err)
	}
	if domainCount != 0 {
		t.Fatalf("expected approval to skip automatic domain creation, got %d domains", domainCount)
	}
	var auditCount int64
	if err := db.Model(&models.AuthzAuditLog{}).Where("action = ?", "reseller_profile_approve").Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit failed: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one audit log, got %d", auditCount)
	}
}

func TestAdminResellerManagementAssignSystemSubdomainCreatesDomainAndAudit(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	c, recorder := newAdminResellerManagementContext(http.MethodPut, fmt.Sprintf("/admin/resellers/profiles/%d/system-domain", profile.ID), strings.NewReader(`{"subdomain":"hello"}`))
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", profile.ID)}}

	adminResellerManagementHTTP(h).AssignSystemDomain(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var domain models.ResellerDomain
	if err := db.Where("reseller_id = ? AND type = ?", profile.ID, models.ResellerDomainTypeSubdomain).First(&domain).Error; err != nil {
		t.Fatalf("expected system domain: %v", err)
	}
	if domain.Domain != "hello.shop.example.test" || domain.Status != models.ResellerDomainStatusActive || domain.VerificationStatus != models.ResellerDomainVerificationVerified || !domain.IsPrimary {
		t.Fatalf("unexpected system domain: %+v", domain)
	}
	var auditCount int64
	if err := db.Model(&models.AuthzAuditLog{}).Where("action = ?", "reseller_profile_system_domain_update").Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit failed: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one audit log, got %d", auditCount)
	}
}

func TestAdminResellerManagementAssignSystemSubdomainReportsMissingSubdomainBase(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	h.ResellerManagementService = resellermodule.NewManagementService(resellerpersistence.NewManagementStore(h.ResellerRepo), config.ResellerConfig{
		Enabled:          true,
		SelfApplyEnabled: true,
	})
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	c, recorder := newAdminResellerManagementContext(http.MethodPut, fmt.Sprintf("/admin/resellers/profiles/%d/system-domain", profile.ID), strings.NewReader(`{"subdomain":"hello"}`))
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", profile.ID)}}

	adminResellerManagementHTTP(h).AssignSystemDomain(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected http 200 envelope, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var resp struct {
		StatusCode int    `json:"status_code"`
		Msg        string `json:"msg"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if resp.StatusCode != response.CodeBadRequest {
		t.Fatalf("expected status_code=400, got %+v body=%s", resp, recorder.Body.String())
	}
	if !strings.Contains(resp.Msg, "分销系统二级域名基础域名未配置") {
		t.Fatalf("expected missing subdomain base message, got %+v body=%s", resp, recorder.Body.String())
	}
}

func TestAdminResellerManagementListProfilesFilters(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusPendingReview)
	activeProfile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	c, recorder := newAdminResellerManagementContext(http.MethodGet, "/admin/resellers/profiles?status=pending_review&page=1&page_size=20", nil)

	adminResellerManagementHTTP(h).ListProfiles(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), fmt.Sprintf(`"id":%d`, profile.ID)) || strings.Contains(recorder.Body.String(), fmt.Sprintf(`"user_id":%d`, activeProfile.UserID)) {
		t.Fatalf("unexpected list body: %s", recorder.Body.String())
	}
}

func TestAdminResellerManagementApproveDomain(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	domain := models.ResellerDomain{ResellerID: profile.ID, Domain: "custom.example.test", Type: models.ResellerDomainTypeCustom, VerificationStatus: models.ResellerDomainVerificationPending, Status: models.ResellerDomainStatusPendingReview}
	if err := db.Create(&domain).Error; err != nil {
		t.Fatalf("create domain failed: %v", err)
	}
	c, recorder := newAdminResellerManagementContext(http.MethodPost, fmt.Sprintf("/admin/resellers/domains/%d/approve", domain.ID), strings.NewReader(`{}`))
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", domain.ID)}}

	adminResellerManagementHTTP(h).ApproveDomain(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var loaded models.ResellerDomain
	if err := db.First(&loaded, domain.ID).Error; err != nil {
		t.Fatalf("load domain failed: %v", err)
	}
	if loaded.Status != models.ResellerDomainStatusActive || loaded.VerificationStatus != models.ResellerDomainVerificationVerified || loaded.VerifiedAt == nil {
		t.Fatalf("unexpected approved domain: %+v", loaded)
	}
}

func TestAdminResellerManagementGetProfileDetailAggregatesOperationalData(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	now := time.Now()
	domains := []models.ResellerDomain{
		{ResellerID: profile.ID, Domain: "r1.shop.example.test", Type: models.ResellerDomainTypeSubdomain, VerificationStatus: models.ResellerDomainVerificationVerified, Status: models.ResellerDomainStatusActive, IsPrimary: true, VerifiedAt: &now},
		{ResellerID: profile.ID, Domain: "custom.example.test", Type: models.ResellerDomainTypeCustom, VerificationStatus: models.ResellerDomainVerificationPending, Status: models.ResellerDomainStatusPendingReview},
	}
	if err := db.Create(&domains).Error; err != nil {
		t.Fatalf("create domains failed: %v", err)
	}
	siteConfig := models.ResellerSiteConfig{ResellerID: profile.ID, SiteName: "运营分销站"}
	if err := db.Create(&siteConfig).Error; err != nil {
		t.Fatalf("create site config failed: %v", err)
	}
	product, skus := seedResellerProductSettingProductForAdminHandler(t, db)
	if _, err := h.ResellerProductSettingRepo.UpsertSetting(models.ResellerProductSetting{
		ResellerID:        profile.ID,
		ProductID:         product.ID,
		SKUID:             skus[0].ID,
		IsListed:          false,
		PricingMode:       models.ResellerPricingModeFixedPrice,
		FixedPriceAmount:  money.FromDecimal(decimal.RequireFromString("129.00")),
		FixedMarkupAmount: money.FromDecimal(decimal.Zero),
	}); err != nil {
		t.Fatalf("upsert setting failed: %v", err)
	}
	order := models.Order{
		OrderNo:              "R202606190001",
		Status:               constants.OrderStatusPaid,
		Currency:             "CNY",
		GuestEmail:           "buyer@example.test",
		TotalAmount:          money.FromDecimal(decimal.RequireFromString("129.00")),
		OriginalAmount:       money.FromDecimal(decimal.RequireFromString("129.00")),
		ResellerID:           &profile.ID,
		ResellerDomain:       "r1.shop.example.test",
		ResellerProfitAmount: money.FromDecimal(decimal.RequireFromString("29.00")),
		PaidAt:               &now,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	orderItem := models.OrderItem{
		OrderID:         order.ID,
		ProductID:       product.ID,
		SKUID:           skus[0].ID,
		Quantity:        1,
		UnitPrice:       money.FromDecimal(decimal.RequireFromString("129.00")),
		TotalPrice:      money.FromDecimal(decimal.RequireFromString("129.00")),
		TitleJSON:       jsonmap.JSON{"zh-CN": "后台商品"},
		SKUSnapshotJSON: jsonmap.JSON{"sku_code": "A"},
		CostPrice:       money.FromDecimal(decimal.RequireFromString("80.00")),
	}
	if err := db.Create(&orderItem).Error; err != nil {
		t.Fatalf("create order item failed: %v", err)
	}
	snapshot := models.ResellerOrderSnapshot{
		OrderID:        order.ID,
		ResellerID:     profile.ID,
		Domain:         "r1.shop.example.test",
		Currency:       "CNY",
		ResellerUserID: profile.UserID,
		BaseAmount:     money.FromDecimal(decimal.RequireFromString("100.00")),
		ResellerAmount: money.FromDecimal(decimal.RequireFromString("129.00")),
		ProfitAmount:   money.FromDecimal(decimal.RequireFromString("29.00")),
		ProfitEligible: true,
	}
	if err := db.Create(&snapshot).Error; err != nil {
		t.Fatalf("create snapshot failed: %v", err)
	}
	ledger := models.ResellerLedgerEntry{ResellerID: profile.ID, OrderID: &order.ID, Type: models.ResellerLedgerTypeOrderProfit, Amount: money.FromDecimal(decimal.RequireFromString("29.00")), Currency: "CNY", IdempotencyKey: "ledger-detail-1", Status: models.ResellerLedgerStatusAvailable}
	if err := db.Create(&ledger).Error; err != nil {
		t.Fatalf("create ledger failed: %v", err)
	}
	balance := models.ResellerBalanceAccount{ResellerID: profile.ID, Currency: "CNY", Status: models.ResellerBalanceStatusNormal, AvailableAmountCache: money.FromDecimal(decimal.RequireFromString("29.00"))}
	if err := db.Create(&balance).Error; err != nil {
		t.Fatalf("create balance failed: %v", err)
	}
	withdraw := models.ResellerWithdrawRequest{ResellerID: profile.ID, Amount: money.FromDecimal(decimal.RequireFromString("10.00")), Currency: "CNY", Channel: "bank", Account: "**** 1234", Status: models.ResellerWithdrawStatusPending}
	if err := db.Create(&withdraw).Error; err != nil {
		t.Fatalf("create withdraw failed: %v", err)
	}

	c, recorder := newAdminResellerManagementContext(http.MethodGet, fmt.Sprintf("/admin/resellers/profiles/%d", profile.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", profile.ID)}}
	adminResellerProfileDetailHTTP(h).GetProfileDetail(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		`"profile"`,
		`"domains"`,
		`"site_config"`,
		`"product_summary"`,
		`"configured_products":1`,
		`"hidden_products":1`,
		`"finance_summary"`,
		`"recent_orders"`,
		`"R202606190001"`,
		`"recent_ledger_entries"`,
		`"recent_withdraws"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected body to contain %s, body=%s", expected, body)
		}
	}
	if strings.Contains(body, "cost_price") || strings.Contains(body, "profit_block_reason") {
		t.Fatalf("detail response leaked internal fields: %s", body)
	}
}

func TestAdminResellerManagementUpdateProfileOperationalConfig(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	body := strings.NewReader(`{"default_markup_percent":"12.50","max_markup_percent":"0","settlement_status":"frozen","reason":"风控复核"}`)
	c, recorder := newAdminResellerManagementContext(http.MethodPut, fmt.Sprintf("/admin/resellers/profiles/%d", profile.ID), body)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", profile.ID)}}

	adminResellerManagementHTTP(h).UpdateProfile(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var loaded models.ResellerProfile
	if err := db.First(&loaded, profile.ID).Error; err != nil {
		t.Fatalf("load profile failed: %v", err)
	}
	if !loaded.DefaultMarkupPercent.Decimal.Equal(decimal.RequireFromString("12.50")) ||
		!loaded.MaxMarkupPercent.Decimal.Equal(decimal.Zero) ||
		loaded.SettlementStatus != models.ResellerSettlementStatusFrozen {
		t.Fatalf("unexpected updated profile: %+v", loaded)
	}
	var auditCount int64
	if err := db.Model(&models.AuthzAuditLog{}).Where("action = ?", "reseller_profile_update").Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit failed: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one audit log, got %d", auditCount)
	}
}

func TestAdminResellerManagementSetPrimaryDomain(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	now := time.Now()
	oldPrimary := models.ResellerDomain{ResellerID: profile.ID, Domain: "r1.shop.example.test", Type: models.ResellerDomainTypeSubdomain, VerificationStatus: models.ResellerDomainVerificationVerified, Status: models.ResellerDomainStatusActive, IsPrimary: true, VerifiedAt: &now}
	nextPrimary := models.ResellerDomain{ResellerID: profile.ID, Domain: "custom.example.test", Type: models.ResellerDomainTypeCustom, VerificationStatus: models.ResellerDomainVerificationVerified, Status: models.ResellerDomainStatusActive, IsPrimary: false, VerifiedAt: &now}
	if err := db.Create(&oldPrimary).Error; err != nil {
		t.Fatalf("create old domain failed: %v", err)
	}
	if err := db.Create(&nextPrimary).Error; err != nil {
		t.Fatalf("create next domain failed: %v", err)
	}

	c, recorder := newAdminResellerManagementContext(http.MethodPost, fmt.Sprintf("/admin/resellers/domains/%d/set-primary", nextPrimary.ID), strings.NewReader(`{}`))
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", nextPrimary.ID)}}
	adminResellerManagementHTTP(h).SetPrimaryDomain(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var loadedOld, loadedNext models.ResellerDomain
	if err := db.First(&loadedOld, oldPrimary.ID).Error; err != nil {
		t.Fatalf("load old domain failed: %v", err)
	}
	if err := db.First(&loadedNext, nextPrimary.ID).Error; err != nil {
		t.Fatalf("load next domain failed: %v", err)
	}
	if loadedOld.IsPrimary || !loadedNext.IsPrimary {
		t.Fatalf("unexpected primary state old=%+v next=%+v", loadedOld, loadedNext)
	}
	var auditCount int64
	if err := db.Model(&models.AuthzAuditLog{}).Where("action = ?", "reseller_domain_set_primary").Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit failed: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one audit log, got %d", auditCount)
	}
}

func TestAdminResellerManagementSetPrimaryDomainRejectsUnverifiedDomain(t *testing.T) {
	h, db := setupAdminResellerManagementHandlerTest(t)
	profile := seedAdminResellerManagementProfile(t, db, models.ResellerProfileStatusActive)
	domain := models.ResellerDomain{ResellerID: profile.ID, Domain: "pending.example.test", Type: models.ResellerDomainTypeCustom, VerificationStatus: models.ResellerDomainVerificationPending, Status: models.ResellerDomainStatusActive}
	if err := db.Create(&domain).Error; err != nil {
		t.Fatalf("create domain failed: %v", err)
	}

	c, recorder := newAdminResellerManagementContext(http.MethodPost, fmt.Sprintf("/admin/resellers/domains/%d/set-primary", domain.ID), strings.NewReader(`{}`))
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", domain.ID)}}
	adminResellerManagementHTTP(h).SetPrimaryDomain(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected wrapped error response 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"status_code":400`) {
		t.Fatalf("expected bad request envelope, body=%s", recorder.Body.String())
	}
}

func seedResellerProductSettingProductForAdminHandler(t *testing.T, db *gorm.DB) (models.Product, []models.ProductSKU) {
	t.Helper()
	category := models.Category{Slug: fmt.Sprintf("admin-setting-cat-%d", time.Now().UnixNano()), NameJSON: jsonmap.JSON{"zh-CN": "分类"}, IsActive: true}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	product := models.Product{
		CategoryID:      category.ID,
		Slug:            fmt.Sprintf("admin-setting-product-%d", time.Now().UnixNano()),
		TitleJSON:       jsonmap.JSON{"zh-CN": "后台商品", "zh-TW": "後台商品", "en-US": "Admin Product"},
		PriceAmount:     money.FromDecimal(decimal.RequireFromString("100.00")),
		CostPriceAmount: money.FromDecimal(decimal.RequireFromString("80.00")),
		IsActive:        true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	skus := []models.ProductSKU{
		{ProductID: product.ID, SKUCode: "A", PriceAmount: money.FromDecimal(decimal.RequireFromString("100.00")), CostPriceAmount: money.FromDecimal(decimal.RequireFromString("80.00")), IsActive: true},
	}
	if err := db.Create(&skus).Error; err != nil {
		t.Fatalf("create skus failed: %v", err)
	}
	return product, skus
}
