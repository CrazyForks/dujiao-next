package gormstore

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
	wallet "github.com/dujiao-next/internal/modules/wallet"
	"github.com/dujiao-next/internal/persistence/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Store is the GORM implementation of the wallet repository port.
type Store struct{ db *gorm.DB }

func New(db *gorm.DB) *Store { return &Store{db: db} }

func (r *Store) WithTx(tx *gorm.DB) *Store {
	if tx == nil {
		return r
	}
	return New(tx)
}

// Transaction remains available for legacy write paths until they migrate.
func (r *Store) Transaction(fn func(tx *gorm.DB) error) error { return r.db.Transaction(fn) }

func (r *Store) GetAccountByUserID(userID uint) (*models.WalletAccount, error) {
	if userID == 0 {
		return nil, nil
	}
	var account models.WalletAccount
	if err := r.db.Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *Store) GetAccountByUserIDForUpdate(userID uint) (*models.WalletAccount, error) {
	if userID == 0 {
		return nil, nil
	}
	var account models.WalletAccount
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *Store) GetAccountsByUserIDs(userIDs []uint) ([]models.WalletAccount, error) {
	if len(userIDs) == 0 {
		return []models.WalletAccount{}, nil
	}
	var accounts []models.WalletAccount
	if err := r.db.Where("user_id IN ?", userIDs).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *Store) CreateAccount(account *models.WalletAccount) error { return r.db.Create(account).Error }
func (r *Store) UpdateAccount(account *models.WalletAccount) error { return r.db.Save(account).Error }

func (r *Store) ListAccounts(filter wallet.AccountListFilter) ([]models.WalletAccount, int64, error) {
	query := r.db.Model(&models.WalletAccount{})
	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var accounts []models.WalletAccount
	if err := gormutil.ApplyPagination(query, filter.Page, filter.PageSize).Order("id desc").Find(&accounts).Error; err != nil {
		return nil, 0, err
	}
	return accounts, total, nil
}

func (r *Store) CreateTransaction(txn *models.WalletTransaction) error { return r.db.Create(txn).Error }
func (r *Store) GetTransactionByReference(reference string) (*models.WalletTransaction, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return nil, nil
	}
	var txn models.WalletTransaction
	if err := r.db.Where("reference = ?", reference).First(&txn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &txn, nil
}

func (r *Store) ListTransactions(filter wallet.TransactionListFilter) ([]models.WalletTransaction, int64, error) {
	query := r.db.Model(&models.WalletTransaction{})
	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.OrderID != 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Direction != "" {
		query = query.Where("direction = ?", filter.Direction)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filter.CreatedTo)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var txns []models.WalletTransaction
	if err := gormutil.ApplyPagination(query, filter.Page, filter.PageSize).Order("id desc").Find(&txns).Error; err != nil {
		return nil, 0, err
	}
	return txns, total, nil
}

func (r *Store) CreateRechargeOrder(order *models.WalletRechargeOrder) error {
	return r.db.Create(order).Error
}
func (r *Store) UpdateRechargeOrder(order *models.WalletRechargeOrder) error {
	return r.db.Save(order).Error
}

func (r *Store) GetRechargeOrderByRechargeNo(userID uint, rechargeNo string) (*models.WalletRechargeOrder, error) {
	if userID == 0 || strings.TrimSpace(rechargeNo) == "" {
		return nil, nil
	}
	var order models.WalletRechargeOrder
	if err := r.db.Where("user_id = ? AND recharge_no = ?", userID, strings.TrimSpace(rechargeNo)).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *Store) GetRechargeOrderByPaymentID(paymentID uint) (*models.WalletRechargeOrder, error) {
	if paymentID == 0 {
		return nil, nil
	}
	var order models.WalletRechargeOrder
	if err := r.db.Where("payment_id = ?", paymentID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *Store) GetRechargeOrderByPaymentIDAndUser(paymentID, userID uint) (*models.WalletRechargeOrder, error) {
	if paymentID == 0 || userID == 0 {
		return nil, nil
	}
	var order models.WalletRechargeOrder
	if err := r.db.Where("payment_id = ? AND user_id = ?", paymentID, userID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *Store) GetRechargeOrderByPaymentIDForUpdate(paymentID uint) (*models.WalletRechargeOrder, error) {
	if paymentID == 0 {
		return nil, nil
	}
	var order models.WalletRechargeOrder
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("payment_id = ?", paymentID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func applyRechargeFilters(query *gorm.DB, filter wallet.RechargeListFilter, includeStatus bool) *gorm.DB {
	if filter.RechargeNo != "" {
		query = query.Where("wallet_recharge_orders.recharge_no LIKE ?", "%"+filter.RechargeNo+"%")
	}
	if filter.UserID != 0 {
		query = query.Where("wallet_recharge_orders.user_id = ?", filter.UserID)
	}
	if filter.UserKeyword != "" {
		like := "%" + filter.UserKeyword + "%"
		query = query.Joins("LEFT JOIN users ON users.id = wallet_recharge_orders.user_id").Where("(users.email LIKE ? OR users.display_name LIKE ?)", like, like)
	}
	if filter.PaymentID != 0 {
		query = query.Where("wallet_recharge_orders.payment_id = ?", filter.PaymentID)
	}
	if filter.ChannelID != 0 {
		query = query.Where("wallet_recharge_orders.channel_id = ?", filter.ChannelID)
	}
	if filter.ProviderType != "" {
		query = query.Where("wallet_recharge_orders.provider_type = ?", filter.ProviderType)
	}
	if filter.ChannelType != "" {
		query = query.Where("wallet_recharge_orders.channel_type = ?", filter.ChannelType)
	}
	if includeStatus && filter.Status != "" {
		query = query.Where("wallet_recharge_orders.status = ?", filter.Status)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("wallet_recharge_orders.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("wallet_recharge_orders.created_at <= ?", *filter.CreatedTo)
	}
	if filter.PaidFrom != nil {
		query = query.Where("wallet_recharge_orders.paid_at >= ?", *filter.PaidFrom)
	}
	if filter.PaidTo != nil {
		query = query.Where("wallet_recharge_orders.paid_at <= ?", *filter.PaidTo)
	}
	return query
}

func (r *Store) ListRechargeOrdersAdmin(filter wallet.RechargeListFilter) ([]models.WalletRechargeOrder, int64, error) {
	query := applyRechargeFilters(r.db.Model(&models.WalletRechargeOrder{}), filter, true)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []models.WalletRechargeOrder
	if err := gormutil.ApplyPagination(query, filter.Page, filter.PageSize).Order("wallet_recharge_orders.id DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *Store) StatsRechargeOrders(filter wallet.RechargeListFilter) (map[string]int64, error) {
	query := r.db.Model(&models.WalletRechargeOrder{})
	if filter.RechargeNo != "" {
		query = query.Where("wallet_recharge_orders.recharge_no LIKE ?", "%"+filter.RechargeNo+"%")
	}
	if filter.UserID != 0 {
		query = query.Where("wallet_recharge_orders.user_id = ?", filter.UserID)
	}
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := query.Select("wallet_recharge_orders.status as status, COUNT(*) as count").Group("wallet_recharge_orders.status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(rows))
	for _, item := range rows {
		result[item.Status] = item.Count
	}
	return result, nil
}

func (r *Store) GetRechargeOrdersByPaymentIDs(paymentIDs []uint) ([]models.WalletRechargeOrder, error) {
	if len(paymentIDs) == 0 {
		return []models.WalletRechargeOrder{}, nil
	}
	var orders []models.WalletRechargeOrder
	if err := r.db.Where("payment_id IN ?", paymentIDs).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
