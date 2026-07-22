package settings

import "github.com/dujiao-next/internal/models"

// Repository 是设置键值持久化的最小端口。
type Repository interface {
	GetByKey(key string) (*models.Setting, error)
	Upsert(key string, value models.JSON) (*models.Setting, error)
}
