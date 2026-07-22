package settings

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

// Repository 是设置键值持久化的最小端口。
type Repository interface {
	GetByKey(key string) (*models.Setting, error)
	Upsert(key string, value jsonmap.JSON) (*models.Setting, error)
}
