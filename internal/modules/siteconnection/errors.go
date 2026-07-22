package siteconnection

import (
	"errors"

	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
)

var (
	// ErrNotFound 与商品映射域共用同一哨兵，保证 errors.Is 跨层一致。
	ErrNotFound = catalogmapping.ErrConnectionNotFound
	ErrInvalid  = errors.New("site connection is invalid")
)
