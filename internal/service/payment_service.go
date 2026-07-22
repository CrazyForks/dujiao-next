package service

import (
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/logger"
	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"
	externalidentitycontract "github.com/dujiao-next/internal/modules/identity/externalidentity/contract"
	usercontract "github.com/dujiao-next/internal/modules/identity/user/contract"
	notificationcontract "github.com/dujiao-next/internal/modules/notification/contract"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/payment/provider"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/repository"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentService 支付服务
type PaymentService struct {
	orderRepo               repository.OrderRepository
	productRepo             paymentProductStore
	productSKURepo          paymentSKUStore
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
	affiliateSvc            AffiliatePaymentLifecycle
	notificationSvc         notificationcontract.NotificationEnqueuer
	procurementSvc          ProcurementCreator
	downstreamCallbackSvc   DownstreamCallbackEnqueuer
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

// DownstreamCallbackEnqueuer 是支付与交付上下文触发下游回调所需的最小端口。
type DownstreamCallbackEnqueuer interface {
	EnqueueCallback(orderID uint)
}

type paymentProductStore interface {
	productcontract.Repository
	BindTx(tx *gorm.DB) productcontract.Repository
}

type paymentSKUStore interface {
	productcontract.SKURepository
	BindTx(tx *gorm.DB) productcontract.SKURepository
}

// AffiliatePaymentLifecycle 是支付成功回调所需的推广返利用例端口。
type AffiliatePaymentLifecycle interface {
	HandleOrderPaid(orderID uint) error
}

// SetProcurementService 设置采购单服务（解决循环依赖）
func (s *PaymentService) SetProcurementService(svc ProcurementCreator) {
	s.procurementSvc = svc
}

// SetDownstreamCallbackService 设置下游回调服务（解决循环依赖）
func (s *PaymentService) SetDownstreamCallbackService(svc DownstreamCallbackEnqueuer) {
	s.downstreamCallbackSvc = svc
}

// SetMemberLevelService 设置会员等级服务
func (s *PaymentService) SetMemberLevelService(svc MemberLevelProgressor) {
	s.memberLevelSvc = svc
}

// PaymentServiceOptions 支付服务构造参数
type PaymentServiceOptions struct {
	OrderRepo                 repository.OrderRepository
	ProductRepo               paymentProductStore
	ProductSKURepo            paymentSKUStore
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
	AffiliateService          AffiliatePaymentLifecycle
	NotificationService       notificationcontract.NotificationEnqueuer
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
