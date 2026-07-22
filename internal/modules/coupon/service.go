package coupon

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"

	"github.com/shopspring/decimal"
)

// CouponService 优惠券服务
type Service struct {
	couponRepo Repository
	usageRepo  UsageRepository
}

type couponEligibility struct {
	subtotal models.Money
	quantity int
}

// NewService 创建优惠券服务
func NewService(couponRepo Repository, usageRepo UsageRepository) *Service {
	return &Service{
		couponRepo: couponRepo,
		usageRepo:  usageRepo,
	}
}

// ApplyCoupon 计算优惠券折扣金额
func (s *Service) ApplyCoupon(subtotal models.Money, code string, userID uint, items []models.OrderItem, isGuest bool, memberLevelID uint) (models.Money, *models.Coupon, error) {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return models.Money{}, nil, ErrInvalid
	}

	coupon, err := s.couponRepo.GetByCode(trimmed)
	if err != nil {
		return models.Money{}, nil, err
	}
	if coupon == nil {
		return models.Money{}, nil, ErrNotFound
	}
	if !coupon.IsActive {
		return models.Money{}, coupon, ErrInactive
	}

	now := time.Now()
	if coupon.StartsAt != nil && now.Before(*coupon.StartsAt) {
		return models.Money{}, coupon, ErrNotStarted
	}
	if coupon.EndsAt != nil && now.After(*coupon.EndsAt) {
		return models.Money{}, coupon, ErrExpired
	}

	if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
		return models.Money{}, coupon, ErrUsageLimit
	}
	if roleErr := resolveCouponPaymentRoleError(coupon, isGuest); roleErr != nil {
		return models.Money{}, coupon, roleErr
	}
	if !matchesCouponMemberLevel(coupon, memberLevelID) {
		return models.Money{}, coupon, ErrMemberLevelNotAllowed
	}

	if coupon.PerUserLimit > 0 && userID != 0 {
		count, err := s.usageRepo.CountByUser(coupon.ID, userID)
		if err != nil {
			return models.Money{}, coupon, err
		}
		if int(count) >= coupon.PerUserLimit {
			return models.Money{}, coupon, ErrPerUserLimit
		}
	}

	eligibility, err := s.resolveCouponEligibility(coupon, items)
	if err != nil {
		return models.Money{}, coupon, err
	}

	if eligibility.subtotal.Decimal.Cmp(coupon.MinAmount.Decimal) < 0 {
		return models.Money{}, coupon, ErrMinAmount
	}

	discount, err := s.calculateDiscount(coupon, eligibility)
	if err != nil {
		return models.Money{}, coupon, err
	}

	if coupon.MaxDiscount.Decimal.GreaterThan(decimal.Zero) && discount.Decimal.GreaterThan(coupon.MaxDiscount.Decimal) {
		discount = models.NewMoneyFromDecimal(coupon.MaxDiscount.Decimal)
	}

	if discount.Decimal.GreaterThan(eligibility.subtotal.Decimal) {
		discount = models.NewMoneyFromDecimal(eligibility.subtotal.Decimal)
	}

	return discount, coupon, nil
}

// matchesCouponRole 判断当前下单角色是否满足优惠券付款角色限制；未配置限制时默认允许。
func matchesCouponRole(coupon *models.Coupon, isGuest bool) bool {
	if coupon == nil || len(coupon.PaymentRoles) == 0 {
		return true
	}
	targetRole := constants.PaymentRoleMember
	if isGuest {
		targetRole = constants.PaymentRoleGuest
	}
	for _, role := range coupon.PaymentRoles {
		if strings.EqualFold(strings.TrimSpace(role), targetRole) {
			return true
		}
	}
	return false
}

// resolveCouponPaymentRoleError 解析付款角色限制不满足时的业务错误。
// 当限制仅单选一个角色时返回更精确的提示错误；否则返回通用角色不匹配错误。
func resolveCouponPaymentRoleError(coupon *models.Coupon, isGuest bool) error {
	if matchesCouponRole(coupon, isGuest) {
		return nil
	}
	if coupon == nil || len(coupon.PaymentRoles) == 0 {
		return ErrPaymentRoleNotAllowed
	}

	roles := make(map[string]struct{}, len(coupon.PaymentRoles))
	for _, role := range coupon.PaymentRoles {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if normalized != constants.PaymentRoleGuest && normalized != constants.PaymentRoleMember {
			continue
		}
		roles[normalized] = struct{}{}
	}

	if len(roles) == 1 {
		if _, ok := roles[constants.PaymentRoleGuest]; ok {
			return ErrPaymentRoleGuestOnly
		}
		if _, ok := roles[constants.PaymentRoleMember]; ok {
			return ErrPaymentRoleMemberOnly
		}
	}
	return ErrPaymentRoleNotAllowed
}

// matchesCouponMemberLevel 判断当前会员等级是否满足优惠券会员等级限制；未配置限制时默认允许。
func matchesCouponMemberLevel(coupon *models.Coupon, memberLevelID uint) bool {
	if coupon == nil || len(coupon.MemberLevels) == 0 {
		return true
	}
	if memberLevelID == 0 {
		return false
	}
	for _, levelID := range coupon.MemberLevels {
		if levelID == memberLevelID {
			return true
		}
	}
	return false
}

func (s *Service) resolveCouponEligibility(coupon *models.Coupon, items []models.OrderItem) (couponEligibility, error) {
	if strings.ToLower(strings.TrimSpace(coupon.ScopeType)) != constants.ScopeTypeProduct {
		return couponEligibility{}, ErrScopeInvalid
	}

	ids, err := DecodeScopeIDs(coupon.ScopeRefIDs)
	if err != nil {
		return couponEligibility{}, ErrScopeInvalid
	}
	if len(ids) == 0 {
		return couponEligibility{}, ErrScopeInvalid
	}

	eligible := decimal.Zero
	eligibleQuantity := 0
	scopeMatched := 0
	wholesaleExcluded := 0
	for _, item := range items {
		if _, ok := ids[item.ProductID]; !ok {
			continue
		}
		scopeMatched++
		if coupon.DisabledWholesalePrice && item.WholesaleDiscount.Decimal.GreaterThan(decimal.Zero) {
			wholesaleExcluded++
			continue
		}
		eligible = eligible.Add(item.TotalPrice.Decimal)
		if item.Quantity > 0 {
			eligibleQuantity += item.Quantity
		}
	}

	if eligible.IsZero() {
		if scopeMatched > 0 && wholesaleExcluded == scopeMatched {
			return couponEligibility{}, ErrWholesaleDisabled
		}
		return couponEligibility{}, ErrScopeInvalid
	}
	return couponEligibility{
		subtotal: models.NewMoneyFromDecimal(eligible),
		quantity: eligibleQuantity,
	}, nil
}

func (s *Service) calculateDiscount(coupon *models.Coupon, eligibility couponEligibility) (models.Money, error) {
	switch strings.ToLower(strings.TrimSpace(coupon.Type)) {
	case constants.CouponTypeFixed:
		if coupon.Value.Decimal.LessThanOrEqual(decimal.Zero) {
			return models.Money{}, ErrInvalid
		}
		if coupon.PerItemDiscount {
			if eligibility.quantity <= 0 {
				return models.Money{}, ErrScopeInvalid
			}
			discount := coupon.Value.Decimal.Mul(decimal.NewFromInt(int64(eligibility.quantity)))
			return models.NewMoneyFromDecimal(discount), nil
		}
		return models.NewMoneyFromDecimal(coupon.Value.Decimal), nil
	case constants.CouponTypePercent:
		if coupon.Value.Decimal.LessThanOrEqual(decimal.Zero) {
			return models.Money{}, ErrInvalid
		}
		percent := coupon.Value.Decimal.Div(decimal.NewFromInt(100))
		discount := eligibility.subtotal.Decimal.Mul(percent)
		return models.NewMoneyFromDecimal(discount), nil
	default:
		return models.Money{}, ErrInvalid
	}
}

// DecodeScopeIDs 解析优惠券适用范围 ID 集合。
func DecodeScopeIDs(raw string) (map[uint]struct{}, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[uint]struct{}{}, nil
	}
	var ids []uint
	if err := json.Unmarshal([]byte(trimmed), &ids); err != nil {
		return nil, err
	}
	result := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		result[id] = struct{}{}
	}
	return result, nil
}
