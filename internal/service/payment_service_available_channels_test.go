package service

import (
	"reflect"
	"sort"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/jsonslice"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestGetAvailableChannelsFilters(t *testing.T) {
	svc, db := setupPaymentServiceWalletTest(t)

	guestOrderInRange := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:               "guest-order-in-range",
		IsActive:           true,
		HideAmountOutRange: true,
		MinAmount:          money.FromDecimal(decimal.RequireFromString("10.00")),
		MaxAmount:          money.FromDecimal(decimal.RequireFromString("100.00")),
		PaymentRoles:       jsonslice.Strings{constants.PaymentRoleGuest},
		PaymentTypes:       jsonslice.Strings{constants.PaymentTypeOrder},
	})
	guestOrderOutOfRange := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:               "guest-order-out-range",
		IsActive:           true,
		HideAmountOutRange: true,
		MinAmount:          money.FromDecimal(decimal.RequireFromString("100.00")),
		MaxAmount:          money.FromDecimal(decimal.RequireFromString("200.00")),
		PaymentRoles:       jsonslice.Strings{constants.PaymentRoleGuest},
		PaymentTypes:       jsonslice.Strings{constants.PaymentTypeOrder},
	})
	memberLv2Wallet := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:         "member-lv2-wallet",
		Icon:         "https://cdn.example.com/icon.png",
		IsActive:     true,
		PaymentRoles: jsonslice.Strings{constants.PaymentRoleMember},
		MemberLevels: jsonslice.Uints{2},
		PaymentTypes: jsonslice.Strings{constants.PaymentTypeWallet},
	})
	memberLv3Wallet := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:         "member-lv3-wallet",
		IsActive:     true,
		PaymentRoles: jsonslice.Strings{constants.PaymentRoleMember},
		MemberLevels: jsonslice.Uints{3},
		PaymentTypes: jsonslice.Strings{constants.PaymentTypeWallet},
	})
	memberLevelOnlyLv2 := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:         "member-level-only-lv2",
		IsActive:     true,
		MemberLevels: jsonslice.Uints{2},
	})
	typeLimitedOrder := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:         "type-limited-order",
		IsActive:     true,
		PaymentTypes: jsonslice.Strings{constants.PaymentTypeOrder},
	})
	unrestricted := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:     "unrestricted",
		IsActive: true,
	})
	inactive := createAvailableChannelFixture(t, db, models.PaymentChannel{
		Name:     "inactive",
		IsActive: true,
	})
	if err := db.Model(&models.PaymentChannel{}).Where("id = ?", inactive.ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("mark inactive channel failed: %v", err)
	}

	amount50 := money.FromDecimal(decimal.RequireFromString("50.00"))
	memberLv2 := &models.User{MemberLevelID: 2}
	memberLv3 := &models.User{MemberLevelID: 3}

	tests := []struct {
		name   string
		filter AvailablePaymentChannelFilter
		want   []uint
	}{
		{
			name: "guest order amount applies range role and payment type filters",
			filter: AvailablePaymentChannelFilter{
				TargetAmount: &amount50,
				PaymentType:  constants.PaymentTypeOrder,
			},
			want: []uint{guestOrderInRange.ID, typeLimitedOrder.ID, unrestricted.ID},
		},
		{
			name: "member lv2 wallet keeps member and level matched channels",
			filter: AvailablePaymentChannelFilter{
				TargetAmount: &amount50,
				User:         memberLv2,
				PaymentType:  constants.PaymentTypeWallet,
			},
			want: []uint{memberLv2Wallet.ID, memberLevelOnlyLv2.ID, unrestricted.ID},
		},
		{
			name: "member lv3 wallet excludes lv2-only channel",
			filter: AvailablePaymentChannelFilter{
				TargetAmount: &amount50,
				User:         memberLv3,
				PaymentType:  constants.PaymentTypeWallet,
			},
			want: []uint{memberLv3Wallet.ID, unrestricted.ID},
		},
		{
			name: "empty payment type keeps type-limited channels for backward compatibility",
			filter: AvailablePaymentChannelFilter{
				PaymentType: "",
			},
			want: []uint{guestOrderInRange.ID, guestOrderOutOfRange.ID, typeLimitedOrder.ID, unrestricted.ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChannels, err := svc.GetAvailableChannels(tt.filter)
			if err != nil {
				t.Fatalf("GetAvailableChannels() error = %v", err)
			}
			gotIDs := collectAvailableChannelIDs(t, gotChannels)
			wantIDs := append([]uint(nil), tt.want...)
			sort.Slice(wantIDs, func(i, j int) bool { return wantIDs[i] < wantIDs[j] })
			if !reflect.DeepEqual(gotIDs, wantIDs) {
				t.Fatalf("channel ids mismatch, got=%v want=%v", gotIDs, wantIDs)
			}
		})
	}

	channels, err := svc.GetAvailableChannels(AvailablePaymentChannelFilter{
		TargetAmount: &amount50,
		User:         memberLv2,
		PaymentType:  constants.PaymentTypeWallet,
	})
	if err != nil {
		t.Fatalf("GetAvailableChannels() error = %v", err)
	}
	byID := indexAvailableChannelsByID(t, channels)
	if _, ok := byID[memberLv2Wallet.ID]["icon"]; !ok {
		t.Fatalf("expected icon field for channel %d", memberLv2Wallet.ID)
	}
	if _, ok := byID[unrestricted.ID]["icon"]; ok {
		t.Fatalf("did not expect icon field for channel %d", unrestricted.ID)
	}
}

func createAvailableChannelFixture(t *testing.T, db *gorm.DB, channel models.PaymentChannel) models.PaymentChannel {
	t.Helper()
	if channel.Name == "" {
		channel.Name = "test-channel"
	}
	if channel.ProviderType == "" {
		channel.ProviderType = constants.PaymentProviderOfficial
	}
	if channel.ChannelType == "" {
		channel.ChannelType = constants.PaymentChannelTypeWechat
	}
	if channel.InteractionMode == "" {
		channel.InteractionMode = constants.PaymentInteractionRedirect
	}
	if channel.FeeRate.Decimal.Equal(decimal.Zero) {
		channel.FeeRate = money.FromDecimal(decimal.Zero)
	}
	if channel.FixedFee.Decimal.Equal(decimal.Zero) {
		channel.FixedFee = money.FromDecimal(decimal.Zero)
	}
	if channel.MinAmount.Decimal.Equal(decimal.Zero) {
		channel.MinAmount = money.FromDecimal(decimal.Zero)
	}
	if channel.MaxAmount.Decimal.Equal(decimal.Zero) {
		channel.MaxAmount = money.FromDecimal(decimal.Zero)
	}
	if !channel.IsActive {
		// 显式传 false 用于测试时，保持原值。
	} else {
		channel.IsActive = true
	}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatalf("create channel failed: %v", err)
	}
	return channel
}

func collectAvailableChannelIDs(t *testing.T, channels []map[string]interface{}) []uint {
	t.Helper()
	ids := make([]uint, 0, len(channels))
	for _, ch := range channels {
		id, ok := ch["id"].(uint)
		if !ok {
			t.Fatalf("channel id type mismatch: %T", ch["id"])
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func indexAvailableChannelsByID(t *testing.T, channels []map[string]interface{}) map[uint]map[string]interface{} {
	t.Helper()
	indexed := make(map[uint]map[string]interface{}, len(channels))
	for _, ch := range channels {
		id, ok := ch["id"].(uint)
		if !ok {
			t.Fatalf("channel id type mismatch: %T", ch["id"])
		}
		indexed[id] = ch
	}
	return indexed
}
