package provider

import (
	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	paymentprovider "github.com/dujiao-next/internal/payment/provider"
	"github.com/dujiao-next/internal/queue"
)

// NewContainer 初始化应用依赖容器。
func NewContainer(cfg *config.Config) *Container {
	if err := cache.InitRedis(&cfg.Redis); err != nil {
		logger.Warnw("provider_init_redis_failed", "error", err)
	}

	var queueClient *queue.Client
	if cfg.Queue.Enabled {
		qc, err := queue.NewClient(&cfg.Queue)
		if err != nil {
			logger.Errorw("provider_init_queue_client_failed", "error", err)
		} else {
			queueClient = qc
		}
	}

	c := &Container{
		Config:                  cfg,
		QueueClient:             queueClient,
		PaymentProviderRegistry: newPaymentProviderRegistry(),
	}
	c.initRepositories()
	c.initServices()
	return c
}

// newPaymentProviderRegistry 注册应用支持的全部支付适配器。
// PaymentService 构造时依赖完整注册表，因此该步骤必须先于 Service 装配。
func newPaymentProviderRegistry() *paymentprovider.Registry {
	registry := paymentprovider.NewRegistry()
	registry.Register(constants.PaymentProviderOfficial, constants.PaymentChannelTypeStripe, paymentprovider.NewStripeAdapter())
	registry.Register(constants.PaymentProviderOfficial, constants.PaymentChannelTypePaypal, paymentprovider.NewPaypalAdapter())
	registry.Register(constants.PaymentProviderOfficial, constants.PaymentChannelTypeWechat, paymentprovider.NewWechatpayAdapter())
	registry.Register(constants.PaymentProviderOfficial, constants.PaymentChannelTypeAlipay, paymentprovider.NewAlipayAdapter())
	registry.Register(constants.PaymentProviderEpay, "", paymentprovider.NewEpayAdapter())
	registry.Register(constants.PaymentProviderEpusdt, "", paymentprovider.NewEpusdtAdapter())
	registry.Register(constants.PaymentProviderBepusdt, "", paymentprovider.NewBepusdtAdapter())
	registry.Register(constants.PaymentProviderDujiaoPay, "", paymentprovider.NewDujiaoPayAdapter())
	registry.Register(constants.PaymentProviderTokenpay, "", paymentprovider.NewTokenpayAdapter())
	registry.Register(constants.PaymentProviderOkpay, "", paymentprovider.NewOkpayAdapter())
	return registry
}
