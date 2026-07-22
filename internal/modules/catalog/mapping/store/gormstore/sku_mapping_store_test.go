package gormstore

import (
	"testing"

	"github.com/dujiao-next/internal/models"
)

func TestSKUMappingStoreListByProductMappingIDs(t *testing.T) {
	_, skuStore, _ := setupMappingStoreTest(t)

	for i, mappingID := range []uint{11, 11, 22} {
		if err := skuStore.Create(&models.SKUMapping{
			ProductMappingID: mappingID,
			LocalSKUID:       uint(100 + i),
			UpstreamSKUID:    uint(200 + i),
		}); err != nil {
			t.Fatalf("create sku mapping failed: %v", err)
		}
	}

	if rows, err := skuStore.ListByProductMappingIDs(nil); err != nil || rows != nil {
		t.Fatalf("empty input should short-circuit: rows=%v err=%v", rows, err)
	}
	rows, err := skuStore.ListByProductMappingIDs([]uint{11})
	if err != nil || len(rows) != 2 {
		t.Fatalf("list by mapping ids: rows=%d err=%v, want 2", len(rows), err)
	}
}

func TestSKUMappingStoreGetByLocalSKUIDReturnsNilWhenMissing(t *testing.T) {
	_, skuStore, _ := setupMappingStoreTest(t)

	mapping, err := skuStore.GetByLocalSKUID(999)
	if err != nil || mapping != nil {
		t.Fatalf("missing sku mapping should be nil without error: mapping=%v err=%v", mapping, err)
	}
}
