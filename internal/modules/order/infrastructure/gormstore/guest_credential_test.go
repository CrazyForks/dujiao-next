package gormstore

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/constants"
	ordercontract "github.com/dujiao-next/internal/modules/order/contract"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

func TestOrderStoreRequiresGuestCredentialSecret(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("New must reject an empty guest credential secret")
		}
	}()
	_ = New(openOrderTenantScopeTestDB(t), "")
}

func TestGuestCredentialIsHashedAtRestAndRawCredentialStillQueries(t *testing.T) {
	db := openOrderTenantScopeTestDB(t)
	repo := New(db, "test-guest-credential-secret-with-32-bytes")
	order := &orderdomain.Order{
		OrderNo:       "GUEST-HASHED",
		GuestEmail:    "guest@example.com",
		GuestPassword: "guest-password",
		Status:        constants.OrderStatusPendingPayment,
		Currency:      "USD",
		TotalAmount:   money.FromDecimal(decimal.NewFromInt(10)),
	}
	if err := repo.WithinTransaction(func(tx ordercontract.Transaction) error {
		return tx.Orders().Create(order, nil)
	}); err != nil {
		t.Fatalf("transactional Create failed: %v", err)
	}

	var stored orderdomain.Order
	if err := db.Select("id", "guest_password").First(&stored, order.ID).Error; err != nil {
		t.Fatalf("reload stored order: %v", err)
	}
	if stored.GuestPassword == "guest-password" || !strings.HasPrefix(stored.GuestPassword, guestCredentialHashPrefix) {
		t.Fatalf("guest credential was not irreversibly hashed at rest: %q", stored.GuestPassword)
	}

	got, err := repo.GetByOrderNoAndGuest("GUEST-HASHED", "guest@example.com", "guest-password")
	if err != nil || got == nil {
		t.Fatalf("raw credential should still authenticate, got order=%v err=%v", got, err)
	}
	got, err = repo.GetByOrderNoAndGuest("GUEST-HASHED", "guest@example.com", "wrong-password")
	if err != nil {
		t.Fatalf("wrong credential query failed unexpectedly: %v", err)
	}
	if got != nil {
		t.Fatalf("wrong credential must not authenticate")
	}
}

func TestGuestCredentialHashLikeRawPasswordCannotBypassHashing(t *testing.T) {
	db := openOrderTenantScopeTestDB(t)
	repo := New(db, "test-guest-credential-secret-with-32-bytes")
	rawPassword := guestCredentialHashPrefix + strings.Repeat("a", 64)
	order := &orderdomain.Order{
		OrderNo:       "GUEST-HASH-LIKE-PASSWORD",
		GuestEmail:    "hash-like@example.com",
		GuestPassword: rawPassword,
		Status:        constants.OrderStatusPendingPayment,
		Currency:      "USD",
		TotalAmount:   money.FromDecimal(decimal.NewFromInt(10)),
	}
	if err := repo.Create(order, nil); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var stored orderdomain.Order
	if err := db.Select("id", "guest_password").First(&stored, order.ID).Error; err != nil {
		t.Fatalf("reload stored order: %v", err)
	}
	if stored.GuestPassword == rawPassword || !isGuestCredentialHash(stored.GuestPassword) {
		t.Fatalf("hash-shaped raw password bypassed hashing: %q", stored.GuestPassword)
	}
	got, err := repo.GetByOrderNoAndGuest(order.OrderNo, order.GuestEmail, rawPassword)
	if err != nil || got == nil {
		t.Fatalf("hash-shaped raw password should authenticate after hashing, got order=%v err=%v", got, err)
	}
}

func TestBackfillGuestCredentialHashesMigratesLegacyPlaintext(t *testing.T) {
	db := openOrderTenantScopeTestDB(t)
	legacy := seedScopedOrder(t, db, "GUEST-LEGACY", 0, "legacy@example.com", "legacy-password", constants.OrderStatusPendingPayment, nil, nil)
	prefixedLegacy := seedScopedOrder(t, db, "GUEST-PREFIXED-LEGACY", 0, "prefixed@example.com", guestCredentialHashPrefix+"legacy-password", constants.OrderStatusPendingPayment, nil, nil)
	repo := New(db, "test-guest-credential-secret-with-32-bytes")

	migrated, err := repo.BackfillGuestCredentialHashes()
	if err != nil {
		t.Fatalf("BackfillGuestCredentialHashes failed: %v", err)
	}
	if migrated != 2 {
		t.Fatalf("migrated = %d, want 2", migrated)
	}

	var stored orderdomain.Order
	if err := db.Select("id", "guest_password").First(&stored, legacy.ID).Error; err != nil {
		t.Fatalf("reload migrated order: %v", err)
	}
	if stored.GuestPassword == "legacy-password" || !strings.HasPrefix(stored.GuestPassword, guestCredentialHashPrefix) {
		t.Fatalf("legacy credential was not hashed: %q", stored.GuestPassword)
	}
	got, err := repo.GetByOrderNoAndGuest("GUEST-LEGACY", "legacy@example.com", "legacy-password")
	if err != nil || got == nil {
		t.Fatalf("migrated credential should authenticate, got order=%v err=%v", got, err)
	}
	got, err = repo.GetByOrderNoAndGuest("GUEST-PREFIXED-LEGACY", "prefixed@example.com", guestCredentialHashPrefix+"legacy-password")
	if err != nil || got == nil {
		t.Fatalf("prefixed legacy credential should be migrated and authenticate, got order=%v err=%v", got, err)
	}
	var prefixedStored orderdomain.Order
	if err := db.Select("id", "guest_password").First(&prefixedStored, prefixedLegacy.ID).Error; err != nil {
		t.Fatalf("reload prefixed migrated order: %v", err)
	}
	if prefixedStored.GuestPassword == guestCredentialHashPrefix+"legacy-password" || !isGuestCredentialHash(prefixedStored.GuestPassword) {
		t.Fatalf("prefixed legacy credential was not hashed: %q", prefixedStored.GuestPassword)
	}
}

func TestBackfillGuestCredentialHashesProcessesBatchesAndSkipsMigratedRows(t *testing.T) {
	db := openOrderTenantScopeTestDB(t)
	repo := New(db, "test-guest-credential-secret-with-32-bytes")
	const plaintextCount = 501
	rows := make([]orderdomain.Order, 0, plaintextCount+1)
	for i := 0; i < plaintextCount; i++ {
		rows = append(rows, orderdomain.Order{
			OrderNo:       fmt.Sprintf("GUEST-BATCH-%04d", i),
			GuestEmail:    fmt.Sprintf("guest-%04d@example.com", i),
			GuestPassword: fmt.Sprintf("plain-%04d", i),
			Status:        constants.OrderStatusPendingPayment,
			Currency:      "USD",
			TotalAmount:   money.FromDecimal(decimal.NewFromInt(10)),
		})
	}
	migratedHash := repo.hashGuestCredential("already@example.com", "already-hashed")
	rows = append(rows, orderdomain.Order{
		OrderNo:       "GUEST-ALREADY-HASHED",
		GuestEmail:    "already@example.com",
		GuestPassword: migratedHash,
		Status:        constants.OrderStatusPendingPayment,
		Currency:      "USD",
		TotalAmount:   money.FromDecimal(decimal.NewFromInt(10)),
	})
	if err := db.CreateInBatches(&rows, 100).Error; err != nil {
		t.Fatalf("seed guest credential batches failed: %v", err)
	}

	migrated, err := repo.BackfillGuestCredentialHashes()
	if err != nil {
		t.Fatalf("BackfillGuestCredentialHashes failed: %v", err)
	}
	if migrated != plaintextCount {
		t.Fatalf("migrated = %d, want %d", migrated, plaintextCount)
	}

	var plaintextRows int64
	if err := db.Model(&orderdomain.Order{}).
		Where("user_id = 0 AND guest_password NOT LIKE ?", guestCredentialHashPrefix+"%").
		Count(&plaintextRows).Error; err != nil {
		t.Fatalf("count plaintext rows failed: %v", err)
	}
	if plaintextRows != 0 {
		t.Fatalf("plaintext rows remaining = %d", plaintextRows)
	}
	var stored orderdomain.Order
	if err := db.Where("order_no = ?", "GUEST-ALREADY-HASHED").First(&stored).Error; err != nil {
		t.Fatalf("reload already migrated order failed: %v", err)
	}
	if stored.GuestPassword != migratedHash {
		t.Fatal("backfill must not rehash an already migrated credential")
	}
}

func TestBackfillGuestCredentialHashesMigratesMalformedHashShapedPlaintext(t *testing.T) {
	db := openOrderTenantScopeTestDB(t)
	repo := New(db, "test-guest-credential-secret-with-32-bytes")
	malformed := guestCredentialHashPrefix + strings.Repeat("z", sha256.Size*2)
	order := orderdomain.Order{
		OrderNo:       "GUEST-MALFORMED-HASH-SHAPE",
		UserID:        0,
		GuestEmail:    "legacy@example.com",
		GuestPassword: malformed,
		Status:        constants.OrderStatusPendingPayment,
		Currency:      "CNY",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("seed malformed guest credential: %v", err)
	}

	migrated, err := repo.BackfillGuestCredentialHashes()
	if err != nil {
		t.Fatalf("BackfillGuestCredentialHashes failed: %v", err)
	}
	if migrated != 1 {
		t.Fatalf("migrated = %d, want 1", migrated)
	}

	var stored orderdomain.Order
	if err := db.First(&stored, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if stored.GuestPassword == malformed {
		t.Fatal("malformed hash-shaped plaintext was not migrated")
	}
	if !isGuestCredentialHash(stored.GuestPassword) {
		t.Fatalf("migrated credential is not a valid hash: %q", stored.GuestPassword)
	}
}
