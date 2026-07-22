package service

import (
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/modules/downstreamcallback"
	externalidentitycontract "github.com/dujiao-next/internal/modules/identity/externalidentity/contract"
	usercontract "github.com/dujiao-next/internal/modules/identity/user/contract"
	"github.com/dujiao-next/internal/modules/notification"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/payment/provider"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/repository"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// PaymentService 支付服务
type PaymentService struct {
	orderRepo               repository.OrderRepository
	productRepo             repository.ProductRepository
	productSKURepo          repository.ProductSKURepository
	paymentRepo             repository.PaymentRepository
	channelRepo             repository.PaymentChannelRepository
	walletRepo              repository.WalletRepository
	userRepo                usercontract.Store
	userOAuthIdentityRepo   externalidentitycontract.Store
	queueClient             *queue.Client
	walletSvc               *WalletService
	settingService          *settingsapp.Service
	defaultEmailConfig      config.EmailConfig
	expireMinutes           int
	affiliateSvc            *AffiliateService
	notificationSvc         *notification.Service
	procurementSvc          ProcurementCreator
	downstreamCallbackSvc   *downstreamcallback.Service
	memberLevelSvc          MemberLevelProgressor
	paymentProviderRegistry *provider.Registry
	resellerAccountingSvc   *ResellerAccountingService
}

type MemberLevelProgressor interface {
	OnOrderPaid(userID uint, amount decimal.Decimal) error
	OnRechargeCompleted(userID uint, amount decimal.Decimal) error
}

type ProcurementCreator interface {
	CreateForOrder(orderID uint) error
}

// SetProcurementService 设置采购单服务（解决循环依赖）
func (s *PaymentService) SetProcurementService(svc ProcurementCreator) {
	s.procurementSvc = svc
}

// SetDownstreamCallbackService 设置下游回调服务（解决循环依赖）
func (s *PaymentService) SetDownstreamCallbackService(svc *downstreamcallback.Service) {
	s.downstreamCallbackSvc = svc
}

// SetMemberLevelService 设置会员等级服务
func (s *PaymentService) SetMemberLevelService(svc MemberLevelProgressor) {
	s.memberLevelSvc = svc
}

// PaymentServiceOptions 支付服务构造参数
type PaymentServiceOptions struct {
	OrderRepo                 repository.OrderRepository
	ProductRepo               repository.ProductRepository
	ProductSKURepo            repository.ProductSKURepository
	PaymentRepo               repository.PaymentRepository
	ChannelRepo               repository.PaymentChannelRepository
	WalletRepo                repository.WalletRepository
	UserStore                 usercontract.Store
	ExternalIdentityStore     externalidentitycontract.Store
	QueueClient               *queue.Client
	WalletService             *WalletService
	SettingService            *settingsapp.Service
	DefaultEmailConfig        config.EmailConfig
	ExpireMinutes             int
	AffiliateService          *AffiliateService
	NotificationService       *notification.Service
	PaymentProviderRegistry   *provider.Registry
	ResellerAccountingService *ResellerAccountingService
}

// NewPaymentService 创建支付服务
func NewPaymentService(opts PaymentServiceOptions) *PaymentService {
	return &PaymentService{
		orderRepo:               opts.OrderRepo,
		productRepo:             opts.ProductRepo,
		productSKURepo:          opts.ProductSKURepo,
		paymentRepo:             opts.PaymentRepo,
		channelRepo:             opts.ChannelRepo,
		walletRepo:              opts.WalletRepo,
		userRepo:                opts.UserStore,
		userOAuthIdentityRepo:   opts.ExternalIdentityStore,
		queueClient:             opts.QueueClient,
		walletSvc:               opts.WalletService,
		settingService:          opts.SettingService,
		defaultEmailConfig:      opts.DefaultEmailConfig,
		expireMinutes:           opts.ExpireMinutes,
		affiliateSvc:            opts.AffiliateService,
		notificationSvc:         opts.NotificationService,
		paymentProviderRegistry: opts.PaymentProviderRegistry,
		resellerAccountingSvc:   opts.ResellerAccountingService,
	}
}

func paymentLogger(kv ...interface{}) *zap.SugaredLogger {
	if len(kv) == 0 {
		return logger.S()
	}
	return logger.SW(kv...)
}
