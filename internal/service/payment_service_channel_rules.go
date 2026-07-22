package service

import (
	"encoding/json"
	"strings"

	"github.com/dujiao-next/internal/models"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/dujiao-next/internal/repository"

	"gorm.io/gorm"
)

// DecodeChannelIDs 解码 JSON 数组字符串 → []uint
func DecodeChannelIDs(raw string) []uint {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "[]" {
		return nil
	}
	var ids []uint
	if err := json.Unmarshal([]byte(trimmed), &ids); err != nil {
		return nil
	}
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			result = append(result, id)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// EncodeChannelIDs 编码 []uint → JSON 数组字符串
func EncodeChannelIDs(ids []uint) string {
	if len(ids) == 0 {
		return ""
	}
	payload, err := json.Marshal(ids)
	if err != nil {
		return ""
	}
	return string(payload)
}

// computeProductChannelIntersection 计算多个商品允许支付渠道的交集
// 空列表表示不限制（全部允许），不参与交集计算
// 返回 nil 表示无限制，返回空切片表示交集为空（无可用渠道）
func computeProductChannelIntersection(products []models.Product) []uint {
	var intersection map[uint]struct{}
	hasRestriction := false

	for _, p := range products {
		allowed := DecodeChannelIDs(p.PaymentChannelIDs)
		if len(allowed) == 0 {
			continue // 该商品不限制
		}
		hasRestriction = true
		allowedSet := make(map[uint]struct{}, len(allowed))
		for _, id := range allowed {
			allowedSet[id] = struct{}{}
		}
		if intersection == nil {
			intersection = allowedSet
		} else {
			for id := range intersection {
				if _, ok := allowedSet[id]; !ok {
					delete(intersection, id)
				}
			}
		}
	}

	if !hasRestriction {
		return nil // 所有商品都不限制
	}

	result := make([]uint, 0, len(intersection))
	for id := range intersection {
		result = append(result, id)
	}
	return result
}

// validateProductPaymentChannel 校验支付渠道是否被订单中的商品允许
// tx 参数可选：在事务内调用时传入 tx 以避免 SQLite 自锁
func (s *PaymentService) validateProductPaymentChannel(items []models.OrderItem, channelID uint, tx ...*gorm.DB) error {
	if len(items) == 0 {
		return nil
	}
	productIDSet := make(map[uint]struct{})
	for _, item := range items {
		if item.ProductID > 0 {
			productIDSet[item.ProductID] = struct{}{}
		}
	}
	if len(productIDSet) == 0 {
		return nil
	}
	productIDs := make([]uint, 0, len(productIDSet))
	for id := range productIDSet {
		productIDs = append(productIDs, id)
	}
	var repo catalogproduct.Repository = s.productRepo
	if len(tx) > 0 && tx[0] != nil {
		repo = s.productRepo.BindTx(tx[0])
	}
	products, err := repo.ListByIDs(productIDs)
	if err != nil {
		return ErrProductFetchFailed
	}
	allowed := computeProductChannelIntersection(products)
	if allowed == nil {
		return nil // 无限制
	}
	for _, id := range allowed {
		if id == channelID {
			return nil
		}
	}
	return ErrPaymentChannelNotAllowedForProduct
}

// validateWalletRechargeChannel 校验支付渠道是否被钱包充值设置允许
func (s *PaymentService) validateWalletRechargeChannel(channelID uint) error {
	if s.settingService == nil {
		return nil
	}
	allowedIDs := s.settingService.GetWalletRechargeChannelIDs()
	if len(allowedIDs) == 0 {
		return nil // 无限制
	}
	for _, id := range allowedIDs {
		if id == channelID {
			return nil
		}
	}
	return ErrPaymentChannelNotAllowedForRecharge
}

// GetAllowedChannelsForProducts 获取商品允许的支付渠道列表
func (s *PaymentService) GetAllowedChannelsForProducts(productIDs []uint) ([]models.PaymentChannel, error) {
	// 仅钱包余额支付模式下不返回任何在线支付渠道
	if s.settingService != nil && s.settingService.GetWalletOnlyPayment() {
		return []models.PaymentChannel{}, nil
	}
	channels, _, err := s.ListChannels(repository.PaymentChannelListFilter{
		Page:       1,
		PageSize:   200,
		ActiveOnly: true,
	})
	if err != nil {
		return nil, err
	}
	if len(productIDs) == 0 {
		return channels, nil
	}
	products, err := s.productRepo.ListByIDs(productIDs)
	if err != nil {
		return nil, ErrProductFetchFailed
	}
	allowed := computeProductChannelIntersection(products)
	if allowed == nil {
		return channels, nil // 无限制
	}
	allowedSet := make(map[uint]struct{}, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = struct{}{}
	}
	filtered := make([]models.PaymentChannel, 0, len(allowed))
	for _, ch := range channels {
		if _, ok := allowedSet[ch.ID]; ok {
			filtered = append(filtered, ch)
		}
	}
	return filtered, nil
}

// GetWalletRechargeChannels 获取钱包充值允许的支付渠道列表
func (s *PaymentService) GetWalletRechargeChannels() ([]models.PaymentChannel, error) {
	channels, _, err := s.ListChannels(repository.PaymentChannelListFilter{
		Page:       1,
		PageSize:   200,
		ActiveOnly: true,
	})
	if err != nil {
		return nil, err
	}
	if s.settingService == nil {
		return channels, nil
	}
	allowedIDs := s.settingService.GetWalletRechargeChannelIDs()
	if len(allowedIDs) == 0 {
		return channels, nil
	}
	allowedSet := make(map[uint]struct{}, len(allowedIDs))
	for _, id := range allowedIDs {
		allowedSet[id] = struct{}{}
	}
	filtered := make([]models.PaymentChannel, 0, len(allowedIDs))
	for _, ch := range channels {
		if _, ok := allowedSet[ch.ID]; ok {
			filtered = append(filtered, ch)
		}
	}
	return filtered, nil
}

// GetAllowedChannelIDsForOrder 获取订单中所有商品允许的支付渠道ID交集
func (s *PaymentService) GetAllowedChannelIDsForOrder(items []models.OrderItem) []uint {
	if len(items) == 0 {
		return nil
	}
	productIDSet := make(map[uint]struct{})
	for _, item := range items {
		if item.ProductID > 0 {
			productIDSet[item.ProductID] = struct{}{}
		}
	}
	if len(productIDSet) == 0 {
		return nil
	}
	productIDs := make([]uint, 0, len(productIDSet))
	for id := range productIDSet {
		productIDs = append(productIDs, id)
	}
	products, err := s.productRepo.ListByIDs(productIDs)
	if err != nil {
		return nil
	}
	return computeProductChannelIntersection(products)
}
