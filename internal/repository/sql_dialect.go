package repository

import (
	"github.com/dujiao-next/internal/persistence/gormutil"

	"gorm.io/gorm"
)

// Legacy Repository 的包内调用暂由这些薄包装转发到共享 GORM 工具；
// 领域 Store 直接依赖 gormutil，避免反向依赖本包。
func dbDialectName(db *gorm.DB) string {
	return gormutil.DialectName(db)
}

func jsonTextExpr(db *gorm.DB, column, key string) string {
	return gormutil.JSONTextExpr(db, column, key)
}

func jsonTextExprByDialect(dialect, column, key string) string {
	return gormutil.JSONTextExprByDialect(dialect, column, key)
}

func jsonArrayLengthExpr(db *gorm.DB, column string) string {
	return gormutil.JSONArrayLengthExpr(db, column)
}

func jsonArrayLengthExprByDialect(dialect, column string) string {
	return gormutil.JSONArrayLengthExprByDialect(dialect, column)
}

func buildLocalizedLikeCondition(db *gorm.DB, plainColumns, jsonColumns []string) (string, int) {
	return gormutil.BuildLocalizedLikeCondition(db, plainColumns, jsonColumns)
}

func buildLocalizedLikeConditionByDialect(dialect string, plainColumns, jsonColumns []string) (string, int) {
	return gormutil.BuildLocalizedLikeConditionByDialect(dialect, plainColumns, jsonColumns)
}

func repeatLikeArgs(like string, count int) []interface{} {
	return gormutil.RepeatLikeArgs(like, count)
}
