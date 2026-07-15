package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupPaymentProviderRenameTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:payment_provider_rename_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	DB = db
	if err := db.AutoMigrate(&PaymentChannel{}, &Setting{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return db
}

func TestEnsurePaymentProviderBepusdtRenameMigration_RenamesAndIsIdempotent(t *testing.T) {
	db := setupPaymentProviderRenameTestDB(t)

	// Seed: 两条 provider_type='epusdt'（旧 BEpusdt）+ 一条无关
	now := time.Now()
	if err := db.Create(&PaymentChannel{
		Name: "old-bepusdt-1", ProviderType: "epusdt", ChannelType: "usdt-trc20",
		InteractionMode: "redirect", IsActive: true, ConfigJSON: JSON{},
		CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed channel 1 failed: %v", err)
	}
	if err := db.Create(&PaymentChannel{
		Name: "old-bepusdt-2", ProviderType: "epusdt", ChannelType: "trx",
		InteractionMode: "redirect", IsActive: true, ConfigJSON: JSON{},
		CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed channel 2 failed: %v", err)
	}
	if err := db.Create(&PaymentChannel{
		Name: "alipay", ProviderType: "official", ChannelType: "alipay",
		InteractionMode: "redirect", IsActive: true, ConfigJSON: JSON{},
		CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed channel 3 failed: %v", err)
	}

	// First run: should rename
	if err := ensurePaymentProviderBepusdtRenameMigration(); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}

	var renamed []PaymentChannel
	if err := db.Where("provider_type = ?", "bepusdt").Find(&renamed).Error; err != nil {
		t.Fatalf("query bepusdt failed: %v", err)
	}
	if len(renamed) != 2 {
		t.Fatalf("expected 2 bepusdt channels after migration, got %d", len(renamed))
	}

	var stillEpusdt int64
	if err := db.Model(&PaymentChannel{}).Where("provider_type = ?", "epusdt").Count(&stillEpusdt).Error; err != nil {
		t.Fatalf("count epusdt failed: %v", err)
	}
	if stillEpusdt != 0 {
		t.Fatalf("expected 0 epusdt channels after migration, got %d", stillEpusdt)
	}

	// Marker should be written
	var marker Setting
	if err := db.First(&marker, "key = ?", "migration/payment_provider_bepusdt_rename_v1").Error; err != nil {
		t.Fatalf("expected marker after migration: %v", err)
	}
	if !migrationDone(marker.ValueJSON) {
		t.Fatalf("marker should have done=true, got %v", marker.ValueJSON)
	}

	// Now seed a NEW real epusdt channel (post-migration scenario)
	if err := db.Create(&PaymentChannel{
		Name: "real-epusdt", ProviderType: "epusdt", ChannelType: "usdt-trc20",
		InteractionMode: "redirect", IsActive: true, ConfigJSON: JSON{},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("seed real epusdt failed: %v", err)
	}

	// Second run: marker hits, should be a no-op for the new real epusdt
	if err := ensurePaymentProviderBepusdtRenameMigration(); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}

	var realEpusdtCount int64
	if err := db.Model(&PaymentChannel{}).Where("name = ? AND provider_type = ?", "real-epusdt", "epusdt").Count(&realEpusdtCount).Error; err != nil {
		t.Fatalf("count real epusdt failed: %v", err)
	}
	if realEpusdtCount != 1 {
		t.Fatalf("idempotency broken: real epusdt should still be provider_type='epusdt', count=%d", realEpusdtCount)
	}

	// And bepusdt count should still be 2 (not 3, the new real epusdt didn't get migrated)
	var bepusdtCount int64
	if err := db.Model(&PaymentChannel{}).Where("provider_type = ?", "bepusdt").Count(&bepusdtCount).Error; err != nil {
		t.Fatalf("count bepusdt failed: %v", err)
	}
	if bepusdtCount != 2 {
		t.Fatalf("idempotency broken: bepusdt count expected 2, got %d", bepusdtCount)
	}
}

func TestEnsurePaymentChannelBepusdtConfigMigration_NormalizesLegacyChannels(t *testing.T) {
	db := setupPaymentProviderRenameTestDB(t)
	tests := []struct {
		name        string
		channelType string
		tradeType   string
	}{
		{name: "legacy-usdt", channelType: "usdt", tradeType: "usdt.trc20"},
		{name: "legacy-usdt-trc20", channelType: "usdt-trc20", tradeType: "usdt.trc20"},
		{name: "legacy-usdc-trc20", channelType: "usdc-trc20", tradeType: "usdt.trc20"},
		{name: "legacy-trx", channelType: "trx", tradeType: "usdt.trc20"},
	}
	for _, tc := range tests {
		if err := db.Create(&PaymentChannel{
			Name: tc.name, ProviderType: "bepusdt", ChannelType: tc.channelType,
			InteractionMode: "redirect", IsActive: true, ConfigJSON: JSON{"gateway_url": "https://bepusdt.example.com"},
		}).Error; err != nil {
			t.Fatalf("seed %s failed: %v", tc.name, err)
		}
	}
	if err := db.Create(&PaymentChannel{
		Name: "explicit", ProviderType: "bepusdt", ChannelType: "usdt-trc20",
		InteractionMode: "redirect", IsActive: true, ConfigJSON: JSON{"trade_type": "usdt.arbitrum"},
	}).Error; err != nil {
		t.Fatalf("seed explicit failed: %v", err)
	}
	if err := db.Create(&PaymentChannel{
		Name: "unknown", ProviderType: "bepusdt", ChannelType: "future-coin",
		InteractionMode: "redirect", IsActive: true, ConfigJSON: JSON{},
	}).Error; err != nil {
		t.Fatalf("seed unknown failed: %v", err)
	}

	if err := ensurePaymentChannelBepusdtConfigMigration(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	for _, tc := range tests {
		var channel PaymentChannel
		if err := db.First(&channel, "name = ?", tc.name).Error; err != nil {
			t.Fatalf("load %s failed: %v", tc.name, err)
		}
		if channel.ChannelType != "bepusdt" {
			t.Fatalf("%s channel_type = %q, want bepusdt", tc.name, channel.ChannelType)
		}
		if got := channel.ConfigJSON["trade_type"]; got != tc.tradeType {
			t.Fatalf("%s trade_type = %v, want %s", tc.name, got, tc.tradeType)
		}
		if got := channel.ConfigJSON["order_mode"]; got != "transaction" {
			t.Fatalf("%s order_mode = %v, want transaction", tc.name, got)
		}
	}

	var explicit PaymentChannel
	if err := db.First(&explicit, "name = ?", "explicit").Error; err != nil {
		t.Fatalf("load explicit failed: %v", err)
	}
	if got := explicit.ConfigJSON["trade_type"]; got != "usdt.arbitrum" {
		t.Fatalf("explicit trade_type changed to %v", got)
	}
	if explicit.ChannelType != "bepusdt" {
		t.Fatalf("explicit channel_type = %q, want bepusdt", explicit.ChannelType)
	}
	if got := explicit.ConfigJSON["order_mode"]; got != "transaction" {
		t.Fatalf("explicit order_mode = %v, want transaction", got)
	}
	var unknown PaymentChannel
	if err := db.First(&unknown, "name = ?", "unknown").Error; err != nil {
		t.Fatalf("load unknown failed: %v", err)
	}
	if unknown.ChannelType != "future-coin" || len(unknown.ConfigJSON) != 0 {
		t.Fatalf("unknown channel should stay unchanged: channel_type=%q config=%v", unknown.ChannelType, unknown.ConfigJSON)
	}

	if err := ensurePaymentChannelBepusdtConfigMigration(); err != nil {
		t.Fatalf("second migration should be idempotent: %v", err)
	}
	var marker Setting
	if err := db.First(&marker, "key = ?", paymentChannelBepusdtConfigMigrationSettingKey).Error; err != nil {
		t.Fatalf("load marker failed: %v", err)
	}
	if !migrationDone(marker.ValueJSON) || marker.ValueJSON["migrated_count"] != float64(len(tests)+1) {
		t.Fatalf("unexpected marker: %v", marker.ValueJSON)
	}
}
