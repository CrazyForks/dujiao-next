package productadmin

import (
	"strconv"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
)

// Delete 删除商品，并在确认无库存和成交记录后级联清理关联数据。
func (s *AdminService) Delete(id string) error {
	product, err := s.products.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return s.errors.NotFound
	}

	// 事务前校验，避免 SQLite 在读后写场景中发生自锁。
	available, err := s.cardSecrets.CountAvailable(product.ID, 0)
	if err != nil {
		return err
	}
	if available > 0 {
		return s.errors.ProductHasStock
	}
	reserved, err := s.cardSecrets.CountReserved(product.ID, 0)
	if err != nil {
		return err
	}
	if reserved > 0 {
		return s.errors.ProductHasStock
	}

	orderCount, err := s.orders.CountOrderItemsByProduct(product.ID)
	if err != nil {
		return err
	}
	if orderCount > 0 {
		return s.errors.ProductHasOrderRecord
	}

	return s.transactions.WithinTransaction(func(repositories DeleteRepositories) error {
		if err := repositories.CardSecrets.DeleteByProduct(product.ID); err != nil {
			return err
		}
		if err := repositories.CardSecretBatches.DeleteByProduct(product.ID); err != nil {
			return err
		}
		if err := repositories.SKUs.DeleteByProduct(product.ID); err != nil {
			return err
		}
		if err := repositories.MemberLevelPrices.DeleteByProduct(product.ID); err != nil {
			return err
		}
		if err := repositories.Carts.DeleteByProduct(product.ID); err != nil {
			return err
		}
		if err := repositories.ProductMappings.DeleteByLocalProduct(product.ID); err != nil {
			return err
		}
		return repositories.Products.Delete(id)
	})
}

// QuickUpdate 快速更新商品部分字段（如 is_active、sort_order）。
func (s *AdminService) QuickUpdate(id string, fields map[string]interface{}) (*productdomain.Product, error) {
	product, err := s.products.GetByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, s.errors.NotFound
	}
	if isActivatingProduct(fields) {
		categoryID := product.CategoryID
		if rawCategoryID, ok := fields["category_id"]; ok {
			parsedCategoryID, parseErr := categoryIDFromValue(rawCategoryID, s.errors.ProductCategoryInvalid)
			if parseErr != nil {
				return nil, parseErr
			}
			categoryID = parsedCategoryID
		}
		if err := validateActivationCategory(s.categories, categoryID, s.errors.ProductCategoryInvalid); err != nil {
			return nil, err
		}
	}
	if err := s.products.QuickUpdate(id, fields); err != nil {
		return nil, err
	}
	return s.products.GetByID(id)
}

// UpdateWholesalePrices 更新商品批发价阶梯，不修改商品其他字段。
func (s *AdminService) UpdateWholesalePrices(id string, inputs []productdomain.WholesalePriceInput) (*productdomain.Product, error) {
	product, err := s.products.GetAdminByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, s.errors.NotFound
	}

	wholesalePrices, err := productdomain.NormalizeWholesalePricesForSKUs(inputs, product.SKUs)
	if err != nil {
		return nil, err
	}
	if err := s.products.QuickUpdate(id, map[string]interface{}{"wholesale_prices": wholesalePrices}); err != nil {
		return nil, err
	}
	return s.products.GetAdminByID(id)
}

func isActivatingProduct(fields map[string]interface{}) bool {
	raw, ok := fields["is_active"]
	if !ok {
		return false
	}
	value, ok := raw.(bool)
	return ok && value
}

func categoryIDFromValue(value interface{}, invalidError error) (uint, error) {
	switch typed := value.(type) {
	case uint:
		return typed, nil
	case uint64:
		return uint(typed), nil
	case uint32:
		return uint(typed), nil
	case int:
		if typed < 0 {
			return 0, invalidError
		}
		return uint(typed), nil
	case int64:
		if typed < 0 {
			return 0, invalidError
		}
		return uint(typed), nil
	case int32:
		if typed < 0 {
			return 0, invalidError
		}
		return uint(typed), nil
	case float64:
		if typed < 0 || typed != float64(uint(typed)) {
			return 0, invalidError
		}
		return uint(typed), nil
	default:
		return 0, invalidError
	}
}

func validateActivationCategory(categoryRepo CategoryRepository, categoryID uint, invalidError error) error {
	if categoryID == 0 || categoryRepo == nil {
		return invalidError
	}

	categoryIDText := strconv.FormatUint(uint64(categoryID), 10)
	category, err := categoryRepo.GetByID(categoryIDText)
	if err != nil {
		return err
	}
	if category == nil || !category.IsActive {
		return invalidError
	}

	childCount, err := categoryRepo.CountChildren(categoryIDText)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return invalidError
	}
	return nil
}
