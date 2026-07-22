package service

import (
	"time"

	usercontract "github.com/dujiao-next/internal/modules/identity/user/contract"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/dujiao-next/internal/repository"
)

const (
	affiliateCodeLength        = 8
	affiliateSplitTypePrefix   = "sp"
	affiliateAttributionWindow = 30 * 24 * time.Hour
	affiliateClickDedupeWindow = 10 * time.Minute
)

// AffiliateService 推广返利业务服务
type AffiliateService struct {
	repo           repository.AffiliateRepository
	userRepo       usercontract.Store
	orderRepo      repository.OrderRepository
	productRepo    catalogproduct.Repository
	settingService *settingsapp.Service
}

// NewAffiliateService 创建推广返利服务
func NewAffiliateService(
	repo repository.AffiliateRepository,
	userRepo usercontract.Store,
	orderRepo repository.OrderRepository,
	productRepo catalogproduct.Repository,
	settingService *settingsapp.Service,
) *AffiliateService {
	return &AffiliateService{
		repo:           repo,
		userRepo:       userRepo,
		orderRepo:      orderRepo,
		productRepo:    productRepo,
		settingService: settingService,
	}
}
