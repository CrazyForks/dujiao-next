package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	affiliatedomain "github.com/dujiao-next/internal/modules/affiliate/domain"
	categorydomain "github.com/dujiao-next/internal/modules/catalog/category/domain"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"

	channelclientdomain "github.com/dujiao-next/internal/modules/channelclient/domain"
	admindomain "github.com/dujiao-next/internal/modules/identity/admin/domain"
	emailverificationdomain "github.com/dujiao-next/internal/modules/identity/emailverification/domain"
	externalidentitydomain "github.com/dujiao-next/internal/modules/identity/externalidentity/domain"
	settingsstore "github.com/dujiao-next/internal/modules/settings/infrastructure/gormstore"
	broadcastdomain "github.com/dujiao-next/internal/modules/telegram/broadcast/domain"
	"github.com/glebarez/sqlite" // 纯 Go SQLite 驱动（基于 modernc.org/sqlite）
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

const (
	manualStockRemainingMigrationSettingKey         = "migration/manual_stock_remaining_v1"
	skuMigrationSettingKey                          = "migration/product_sku_v1"
	categoryParentMigrationSettingKey               = "migration/category_parent_v1"
	paymentProviderBepusdtRenameMigrationSettingKey = "migration/payment_provider_bepusdt_rename_v1"
	paymentChannelBepusdtConfigMigrationSettingKey  = "migration/payment_channel_bepusdt_config_v2"
	orderItemOriginalPriceMigrationKey              = "migration/order_item_original_price_v1"
	manualStockUnlimitedValue                       = -1
)

// DBPoolConfig 数据库连接池配置
type DBPoolConfig struct {
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
	ConnMaxIdleTimeSeconds int
}

// InitDB 初始化数据库连接
func InitDB(driver, dsn string, pool DBPoolConfig, mode string) error {
	var err error
	normalized := strings.ToLower(strings.TrimSpace(driver))
	var dialector gorm.Dialector
	switch normalized {
	case "", "sqlite":
		// glebarez/sqlite 是基于 modernc.org/sqlite 的纯 Go 驱动
		// 追加 PRAGMA 参数避免并发写入时 busy-spin 导致 CPU 飙升
		dialector = sqlite.Open(appendSQLitePragmas(dsn))
	case "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger:  newGORMLogger(mode, nil),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	applyDBPool(sqlDB, pool)

	// SQLite 额外执行 PRAGMA 确保 WAL 模式生效
	if normalized == "" || normalized == "sqlite" {
		DB.Exec("PRAGMA journal_mode=WAL")
		DB.Exec("PRAGMA busy_timeout=5000")
		DB.Exec("PRAGMA synchronous=NORMAL")
	}
	return nil
}

func newGORMLogger(mode string, writer gormlogger.Writer) gormlogger.Interface {
	if !strings.EqualFold(strings.TrimSpace(mode), "release") {
		return gormlogger.Default.LogMode(gormlogger.Info)
	}
	if writer == nil {
		writer = log.New(os.Stdout, "\r\n", log.LstdFlags)
	}
	return gormlogger.New(writer, gormlogger.Config{
		SlowThreshold:             2 * time.Second,
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  false,
	})
}

// appendSQLitePragmas 在 SQLite DSN 中追加关键 PRAGMA 参数
func appendSQLitePragmas(dsn string) string {
	// glebarez/sqlite 支持 ?_pragma=key=value 形式
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep +
		"_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)"
}

func applyDBPool(sqlDB *sql.DB, pool DBPoolConfig) {
	if sqlDB == nil {
		return
	}
	if pool.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	}
	if pool.MaxIdleConns >= 0 {
		sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	}
	if pool.ConnMaxLifetimeSeconds > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(pool.ConnMaxLifetimeSeconds) * time.Second)
	}
	if pool.ConnMaxIdleTimeSeconds > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(pool.ConnMaxIdleTimeSeconds) * time.Second)
	}
}

// AutoMigrate 自动迁移所有数据库表
func AutoMigrate() error {
	if err := DB.AutoMigrate(
		&admindomain.Admin{},
		&userdomain.User{},
		&externalidentitydomain.Identity{},
		&affiliatedomain.Profile{},
		&affiliatedomain.Click{},
		&affiliatedomain.Commission{},
		&affiliatedomain.WithdrawRequest{},
		&ResellerProfile{},
		&ResellerDomain{},
		&ResellerSiteConfig{},
		&ResellerProductSetting{},
		&ResellerOrderSnapshot{},
		&ResellerLedgerEntry{},
		&ResellerWithdrawRequest{},
		&ResellerBalanceAccount{},
		&ResellerRelatedAccount{},
		&WalletAccount{},
		&WalletTransaction{},
		&WalletRechargeOrder{},
		&UserLoginLog{},
		&AuthzAuditLog{},
		&NotificationLog{},
		&AdminLoginLog{},
		&emailverificationdomain.Code{},
		&Order{},
		&OrderItem{},
		&OrderRefundRecord{},
		&CartItem{},
		&PaymentChannel{},
		&Payment{},
		&CardSecret{},
		&CardSecretBatch{},
		&GiftCard{},
		&GiftCardBatch{},
		&Fulfillment{},
		&Coupon{},
		&CouponUsage{},
		&Promotion{},
		&categorydomain.Category{},
		&productdomain.Product{},
		&productdomain.ProductSKU{},
		&Post{},
		&PostProduct{},
		&PostCategory{},
		&Banner{},
		&settingsstore.SettingRecord{},
		&ApiCredential{},
		&SiteConnection{},
		&ProductMapping{},
		&SKUMapping{},
		&ProcurementOrder{},
		&DownstreamOrderRef{},
		&ReconciliationJob{},
		&ReconciliationItem{},
		&channelclientdomain.Client{},
		&broadcastdomain.Broadcast{},
		&MemberLevel{},
		&MemberLevelPrice{},
		&Media{},
	); err != nil {
		return err
	}

	if err := migrateCartSKUUniqueIndex(); err != nil {
		return err
	}

	if err := ensureProductSKUMigration(); err != nil {
		return err
	}
	if err := ensureManualStockRemainingMigration(); err != nil {
		return err
	}
	if err := ensureCategoryParentMigration(); err != nil {
		return err
	}
	if err := ensurePaymentProviderBepusdtRenameMigration(); err != nil {
		return err
	}
	if err := ensurePaymentChannelBepusdtConfigMigration(); err != nil {
		return err
	}
	if err := ensureOrderItemOriginalPriceMigration(); err != nil {
		return err
	}
	if err := ensureResellerIndexes(DB); err != nil {
		return err
	}

	// 移除历史遗留商品币种列，统一由站点配置提供币种。
	if DB.Migrator().HasColumn(&productdomain.Product{}, "price_currency") {
		if err := DB.Migrator().DropColumn(&productdomain.Product{}, "price_currency"); err != nil {
			return err
		}
	}
	return nil
}
