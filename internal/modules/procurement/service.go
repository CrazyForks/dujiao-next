package procurement

import (
	"errors"
	"time"

	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/notification"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/upstream"

	"github.com/hibiken/asynq"
)

var (
	ErrNotFound           = errors.New("procurement order not found")
	ErrExists             = errors.New("procurement order already exists")
	ErrStatusInvalid      = errors.New("procurement order status invalid")
	ErrOrderNotFound      = errors.New("order not found")
	ErrConnectionNotFound = errors.New("site connection not found")
)

// ListFilter 描述采购单管理列表与状态统计使用的筛选条件。
type ListFilter struct {
	ConnectionID    uint
	Status          string
	LocalOrderNo    string
	UpstreamOrderNo string
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	Page            int
	PageSize        int
}

// Repository 是采购模块所需的持久化端口。
// 对账服务也复用最后一个按连接和时间范围查询的方法。
type Repository interface {
	GetByID(id uint) (*models.ProcurementOrder, error)
	GetByLocalOrderID(localOrderID uint) (*models.ProcurementOrder, error)
	GetByLocalOrderNo(localOrderNo string) (*models.ProcurementOrder, error)
	Create(order *models.ProcurementOrder) error
	UpdateStatus(id uint, status string, updates map[string]interface{}) error
	List(filter ListFilter) ([]models.ProcurementOrder, int64, error)
	StatsByStatus(filter ListFilter) (map[string]int64, error)
	ListByConnectionAndTimeRange(connectionID uint, start, end time.Time) ([]models.ProcurementOrder, error)
}

type OrderRepository interface {
	GetByID(id uint) (*models.Order, error)
	GetByIDs(ids []uint) ([]models.Order, error)
	UpdateStatus(id uint, status string, updates map[string]interface{}) error
}

type ProductMappingRepository interface {
	GetByLocalProductID(productID uint) (*models.ProductMapping, error)
}

type SKUMappingRepository interface {
	GetByLocalSKUID(skuID uint) (*models.SKUMapping, error)
}

// ConnectionProvider 隔离站点连接的读取与上游协议适配器构造。
type ConnectionProvider interface {
	GetByID(id uint) (*siteconnectiondomain.Connection, error)
	GetAdapter(conn *siteconnectiondomain.Connection) (upstream.Adapter, error)
}

// Enqueuer 是采购模块真正使用到的队列能力。
type Enqueuer interface {
	EnqueueProcurementSubmit(payload queue.ProcurementSubmitPayload, opts ...asynq.Option) error
	EnqueueProcurementPollStatus(payload queue.ProcurementPollStatusPayload, delay time.Duration) error
}

// OrderLifecycle 收口仍属于订单领域的事务、父订单聚合与邮件策略。
// 采购模块只表达何时发生，不复制订单领域的实现规则。
type OrderLifecycle interface {
	CreateUpstreamFulfillment(orderID uint, fulfillment *upstream.UpstreamFulfillment, now time.Time) error
	SyncParentStatus(parentID uint, now time.Time) (string, error)
	EnqueueStatusEmail(orderID uint, status string) (skipped bool, err error)
}

type DownstreamCallbackEnqueuer interface {
	EnqueueCallback(orderID uint)
}

type BotFulfillmentNotifier interface {
	NotifyBotOrderFulfilled(userID, orderID uint)
}

type NotificationEnqueuer interface {
	Enqueue(input notification.EnqueueInput) error
}

// ServiceOptions 显式声明采购模块的依赖，避免继续扩张位置参数构造器。
type ServiceOptions struct {
	Repository         Repository
	Orders             OrderRepository
	ProductMappings    ProductMappingRepository
	SKUMappings        SKUMappingRepository
	Connections        ConnectionProvider
	Queue              Enqueuer
	OrderLifecycle     OrderLifecycle
	DownstreamCallback DownstreamCallbackEnqueuer
	BotNotifier        BotFulfillmentNotifier
	Notifications      NotificationEnqueuer
}

// Service 负责采购单从创建、提交、轮询到交付/退款回调的完整生命周期。
type Service struct {
	procRepo           Repository
	orderRepo          OrderRepository
	mappingRepo        ProductMappingRepository
	skuMapRepo         SKUMappingRepository
	connections        ConnectionProvider
	queue              Enqueuer
	orderLifecycle     OrderLifecycle
	downstreamCallback DownstreamCallbackEnqueuer
	botNotifier        BotFulfillmentNotifier
	notifications      NotificationEnqueuer
}

func NewService(options ServiceOptions) *Service {
	return &Service{
		procRepo:           options.Repository,
		orderRepo:          options.Orders,
		mappingRepo:        options.ProductMappings,
		skuMapRepo:         options.SKUMappings,
		connections:        options.Connections,
		queue:              options.Queue,
		orderLifecycle:     options.OrderLifecycle,
		downstreamCallback: options.DownstreamCallback,
		botNotifier:        options.BotNotifier,
		notifications:      options.Notifications,
	}
}
