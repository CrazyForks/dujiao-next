// Package reseller contains persistence adapters for the reseller module.
//
// The adapters isolate the module from the legacy repository interfaces while
// those repositories still own the GORM transaction boundary.
package reseller

import (
	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/repository"
	"gorm.io/gorm"
)

// NewManagementStore adapts the legacy reseller repository to the focused
// management store required by modules/reseller.
func NewManagementStore(repo repository.ResellerRepository) resellermodule.ManagementStore {
	return &managementStore{repo: repo}
}

type managementStore struct {
	repo repository.ResellerRepository
}

func (s *managementStore) WithinTransaction(fn func(resellermodule.ManagementStore) error) error {
	if s == nil || s.repo == nil {
		return fn(s)
	}
	return s.repo.Transaction(func(tx *gorm.DB) error {
		return fn(&managementStore{repo: s.repo.WithTx(tx)})
	})
}

func (s *managementStore) GetProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	return s.repo.GetProfileByUserID(userID)
}

func (s *managementStore) GetProfileByID(id uint) (*models.ResellerProfile, error) {
	return s.repo.GetProfileByID(id)
}

func (s *managementStore) CreateProfile(profile *models.ResellerProfile) error {
	return s.repo.CreateProfile(profile)
}

func (s *managementStore) UpdateProfile(profile *models.ResellerProfile) error {
	return s.repo.UpdateProfile(profile)
}

func (s *managementStore) ListDomainsByResellerID(resellerID uint) ([]models.ResellerDomain, error) {
	return s.repo.ListDomainsByResellerID(resellerID)
}

func (s *managementStore) UpsertDomain(domain models.ResellerDomain) (*models.ResellerDomain, error) {
	return s.repo.UpsertDomain(domain)
}

func (s *managementStore) GetDomainByID(id uint) (*models.ResellerDomain, error) {
	return s.repo.GetDomainByID(id)
}

func (s *managementStore) GetDomainByIDForUpdate(id uint) (*models.ResellerDomain, error) {
	return s.repo.GetDomainByIDForUpdate(id)
}

func (s *managementStore) UpdateDomain(domain *models.ResellerDomain) error {
	return s.repo.UpdateDomain(domain)
}

func (s *managementStore) FindDomainByHost(host string) (*models.ResellerDomain, error) {
	return s.repo.FindDomainByHost(host)
}

// NewProductSettingStore adapts the split product-setting and reseller
// repositories to the focused store required by modules/reseller.
func NewProductSettingStore(
	settingRepo repository.ResellerProductSettingRepository,
	resellerRepo repository.ResellerRepository,
) resellermodule.ProductSettingStore {
	return &productSettingStore{settingRepo: settingRepo, resellerRepo: resellerRepo}
}

type productSettingStore struct {
	settingRepo  repository.ResellerProductSettingRepository
	resellerRepo repository.ResellerRepository
}

func (s *productSettingStore) WithinTransaction(fn func(resellermodule.ProductSettingStore) error) error {
	if s == nil || s.settingRepo == nil {
		return fn(s)
	}
	return s.settingRepo.Transaction(func(tx *gorm.DB) error {
		return fn(&productSettingStore{
			settingRepo:  s.settingRepo.WithTx(tx),
			resellerRepo: s.resellerRepo,
		})
	})
}

func (s *productSettingStore) GetProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	return s.resellerRepo.GetProfileByUserID(userID)
}

func (s *productSettingStore) GetProfileByID(id uint) (*models.ResellerProfile, error) {
	return s.resellerRepo.GetProfileByID(id)
}

func (s *productSettingStore) ListProductsWithSettings(filter resellermodule.ProductSettingListFilter) ([]resellermodule.ProductSettingProductRow, int64, error) {
	rows, total, err := s.settingRepo.ListProductsWithSettings(repository.ResellerProductSettingListFilter{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		ResellerID: filter.ResellerID,
		CategoryID: filter.CategoryID,
		Keyword:    filter.Keyword,
		Configured: filter.Configured,
		Listed:     filter.Listed,
		OnlyActive: filter.OnlyActive,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]resellermodule.ProductSettingProductRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, resellermodule.ProductSettingProductRow{Product: row.Product, Settings: row.Settings})
	}
	return out, total, nil
}

func (s *productSettingStore) GetProductWithSettings(resellerID, productID uint) (*resellermodule.ProductSettingProductRow, error) {
	row, err := s.settingRepo.GetProductWithSettings(resellerID, productID)
	if err != nil || row == nil {
		return nil, err
	}
	return &resellermodule.ProductSettingProductRow{Product: row.Product, Settings: row.Settings}, nil
}

func (s *productSettingStore) UpsertSetting(setting models.ResellerProductSetting) (*models.ResellerProductSetting, error) {
	return s.settingRepo.UpsertSetting(setting)
}

func (s *productSettingStore) DeleteSetting(resellerID, productID, skuID uint) error {
	return s.settingRepo.DeleteSetting(resellerID, productID, skuID)
}

func (s *productSettingStore) ListAdminSettings(filter resellermodule.ProductSettingAdminListFilter) ([]models.ResellerProductSetting, int64, error) {
	return s.settingRepo.ListAdminSettings(repository.ResellerProductSettingAdminListFilter{
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		ResellerID:  filter.ResellerID,
		UserID:      filter.UserID,
		ProductID:   filter.ProductID,
		Keyword:     filter.Keyword,
		PricingMode: filter.PricingMode,
		Configured:  filter.Configured,
		Listed:      filter.Listed,
	})
}

func (s *productSettingStore) SummarizeByResellerID(resellerID uint) (resellermodule.ProductSettingSummary, error) {
	summary, err := s.settingRepo.SummarizeByResellerID(resellerID)
	if err != nil {
		return resellermodule.ProductSettingSummary{}, err
	}
	return resellermodule.ProductSettingSummary{
		ConfiguredProducts: summary.ConfiguredProducts,
		HiddenProducts:     summary.HiddenProducts,
		SKUOverrides:       summary.SKUOverrides,
		PricingOverrides:   summary.PricingOverrides,
	}, nil
}
