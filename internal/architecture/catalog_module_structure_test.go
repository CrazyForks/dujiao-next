package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCatalogStockConsumersUseSharedPolicy(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	consumers := []string{
		"internal/transport/http/catalog/public_view.go",
		"internal/transport/http/catalog/public_stock.go",
		"internal/transport/http/channel/channel_catalog.go",
		"internal/transport/http/upstream/upstream_catalog.go",
	}
	forbiddenFunctions := map[string]struct{}{
		"normalizePublicStockDisplayMode":    {},
		"buildPublicStockDisplay":            {},
		"normalizePublicStockStatus":         {},
		"publicStockRange":                   {},
		"maskPublicStockInt":                 {},
		"maskPublicStockInt64":               {},
		"maskPublicStockSold":                {},
		"computeStockCount":                  {},
		"normalizeChannelStockDisplayMode":   {},
		"buildChannelStockDisplay":           {},
		"normalizeChannelStockDisplayStatus": {},
		"channelStockRange":                  {},
		"computeStockStatus":                 {},
	}

	for _, relativePath := range consumers {
		relativePath := relativePath
		t.Run(filepath.Base(relativePath), func(t *testing.T) {
			path := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", relativePath, err)
			}

			importsCatalog := false
			for _, imported := range parsed.Imports {
				importPath, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					t.Fatalf("unquote import in %s: %v", relativePath, err)
				}
				if importPath == moduleImportPath+"/internal/modules/catalog" {
					importsCatalog = true
				}
			}
			if !importsCatalog {
				t.Errorf("%s must consume the shared catalog stock policy", relativePath)
			}

			for _, declaration := range parsed.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if _, forbidden := forbiddenFunctions[function.Name.Name]; forbidden {
					t.Errorf("%s redeclares legacy stock policy helper %s", relativePath, function.Name.Name)
				}
			}
		})
	}
}

func TestCatalogCategoryImplementationLivesInBoundedContextDirectories(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "catalog")
	storeRoot := filepath.Join(moduleRoot, "store", "gormstore")
	transportRoot := filepath.Join(repositoryRoot, "internal", "transport", "http", "catalog")
	integrationRoot := filepath.Join(repositoryRoot, "internal", "integration", "catalog")

	assertFileDeclaresTypes(t, filepath.Join(moduleRoot, "category_service.go"), []string{
		"CategoryRepository", "CategoryService", "CreateCategoryInput",
	})
	assertFileDeclaresFunctions(t, filepath.Join(moduleRoot, "category_service.go"), []string{
		"NewCategoryService",
	})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "category_store.go"), []string{"CategoryStore"})
	assertFileDeclaresFunctions(t, filepath.Join(storeRoot, "category_store.go"), []string{"NewCategoryStore"})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_category_handler.go"), []string{
		"CategoryService", "AdminCategoryHandler", "CreateCategoryRequest", "PatchCategoryActiveRequest",
	})
	assertFileDeclaresFunctions(t, filepath.Join(transportRoot, "routes.go"), []string{
		"RegisterPublicRoutes",
		"RegisterAdminCategoryRoutes",
		"RegisterAdminProductRoutes",
		"RegisterAdminProductMappingRoutes",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "public_handler.go"), []string{
		"PublicProductQueries", "PublicCategoryQueries", "ResellerDisplayPricer",
		"ProductPromotionDecorator", "MemberLevelPricing", "LocalProductMappingReader",
		"RelatedPostReader", "PublicHandler",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_product_handler.go"), []string{
		"ProductQueries", "ProductCommands", "LowStockThresholdProvider",
		"ProductMappingLookup", "SKUMappingLookup", "AdminProductHandler",
	})
	assertFileDeclaresTypes(t, filepath.Join(transportRoot, "admin_product_mapping_handler.go"), []string{
		"ProductMappingService", "AdminProductMappingHandler",
	})

	assertDirectoryGoFileBudget(t, moduleRoot, 4)
	assertDirectoryGoFileBudget(t, storeRoot, 2)
	assertDirectoryGoFileBudget(t, transportRoot, 14)
	assertDirectoryGoFileBudget(t, integrationRoot, 2)

	for _, legacy := range []string{
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "catalog.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "catalog_view.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "public", "catalog_stock.go"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy public catalog handler must stay removed: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy public catalog handler: %v", err)
		}
	}
}

func TestCatalogCategoryLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	patterns := []string{
		filepath.Join(repositoryRoot, "internal", "service", "category_service*.go"),
		filepath.Join(repositoryRoot, "internal", "repository", "category_repository*.go"),
		filepath.Join(repositoryRoot, "internal", "http", "handlers", "admin", "admin_category*.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		if len(matches) != 0 {
			t.Errorf("legacy Catalog category files must stay removed: %v", matches)
		}
	}
}

func TestCatalogProductImplementationLivesInNestedBoundedContext(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	productRoot := filepath.Join(repositoryRoot, "internal", "modules", "catalog", "product")
	applicationRoot := filepath.Join(productRoot, "application")
	adminRoot := filepath.Join(applicationRoot, "admin")
	writeRoot := filepath.Join(applicationRoot, "write")
	domainRoot := filepath.Join(productRoot, "domain")
	manualFormRoot := filepath.Join(productRoot, "manualform")
	storeRoot := filepath.Join(productRoot, "store", "gormstore")
	sharedGORMRoot := filepath.Join(repositoryRoot, "internal", "persistence", "gormutil")

	assertFileDeclaresTypes(t, filepath.Join(productRoot, "ports.go"), []string{
		"ListFilter", "Repository", "SKURepository",
	})
	if _, err := os.Stat(filepath.Join(productRoot, "errors.go")); err != nil {
		t.Fatalf("catalog product errors.go must exist: %v", err)
	}
	assertFileDeclaresTypes(t, filepath.Join(applicationRoot, "query.go"), []string{
		"ProductRepository", "CategoryRepository", "HiddenProductRepository", "StockCounter",
		"TenantContext", "Options", "Service",
	})
	assertFileDeclaresFunctions(t, filepath.Join(applicationRoot, "query.go"), []string{
		"NewService", "ListPublic", "ListPublicForTenant", "ListForUpstreamSync", "ListPublicExact",
		"GetPublicBySlug", "GetPublicBySlugForTenant", "ListAdmin", "GetAdminByID",
	})
	assertFileDeclaresFunctions(t, filepath.Join(applicationRoot, "stock.go"), []string{
		"ApplyAutoStockCounts", "resolveLegacyStockTargetSKUIndex",
	})
	assertFileDeclaresTypes(t, filepath.Join(adminRoot, "service.go"), []string{
		"ProductRepository", "CategoryRepository", "CardSecretStockRepository", "OrderHistoryRepository",
		"ProductDeleteRepository", "CardSecretDeleteRepository", "CardSecretBatchDeleteRepository",
		"SKUDeleteRepository", "MemberLevelPriceDeleteRepository", "CartDeleteRepository",
		"ProductMappingDeleteRepository", "DeleteRepositories", "UnitOfWork", "ErrorSet", "Options", "AdminService",
	})
	assertFileDeclaresFunctions(t, filepath.Join(adminRoot, "service.go"), []string{"NewAdminService"})
	assertFileDeclaresFunctions(t, filepath.Join(adminRoot, "operations.go"), []string{
		"Delete", "QuickUpdate", "UpdateWholesalePrices", "isActivatingProduct",
		"categoryIDFromValue", "validateActivationCategory",
	})
	assertFileDeclaresFunctions(t, filepath.Join(adminRoot, "operations_test.go"), []string{
		"TestAdminServiceDeleteStopsBeforeTransactionWhenStockExists",
		"TestAdminServiceDeleteUsesAllCascadePorts",
		"TestAdminServiceQuickUpdateValidatesActivationCategory",
		"TestAdminServiceUpdateWholesalePricesCanonicalizesSKU",
	})
	assertFileDeclaresTypes(t, filepath.Join(writeRoot, "service.go"), []string{
		"ProductRepository", "SKURepository", "CategoryRepository", "PaymentChannelRepository",
		"CardSecretStockRepository", "TransactionRepositories", "UnitOfWork", "ErrorSet", "Options",
		"WriteService", "CreateProductInput", "WholesalePriceInput", "ProductSKUInput",
	})
	assertFileDeclaresFunctions(t, filepath.Join(writeRoot, "service.go"), []string{"NewWriteService"})
	assertFileDeclaresFunctions(t, filepath.Join(writeRoot, "create.go"), []string{"Create"})
	assertFileDeclaresFunctions(t, filepath.Join(writeRoot, "update.go"), []string{"Update"})
	assertFileDeclaresFunctions(t, filepath.Join(writeRoot, "skus.go"), []string{
		"syncSingleProductSKU", "pickSingleModeTargetSKUIndex", "normalizeProductSKUInputs",
		"minActiveCostPrice", "applyProductSKUsWithStockGuard", "ensureAutoSKUCardSecretStockSafe",
	})
	assertFileDeclaresFunctions(t, filepath.Join(writeRoot, "skus_test.go"), []string{
		"TestSyncSingleProductSKUMultipleRowsKeepsSingleActive",
		"TestSyncSingleProductSKUNoActivePrefersDefaultCode",
	})
	assertFileDeclaresTypes(t, filepath.Join(domainRoot, "pricing.go"), []string{"WholesalePriceInput"})
	assertFileDeclaresFunctions(t, filepath.Join(domainRoot, "pricing.go"), []string{
		"NormalizeWholesalePrices", "NormalizeWholesalePricesForSKUs",
		"ResolveWholesaleUnitPrice", "ResolveWholesaleUnitPriceWithMatchQuantity", "ResolveWholesaleUnitPriceForSKU",
	})
	assertFileDeclaresFunctions(t, filepath.Join(domainRoot, "purchase_policy.go"), []string{
		"NormalizePurchaseType", "NormalizeFulfillmentType", "NormalizeStockDisplayMode",
		"ValidateCategoryAssignment", "NormalizePurchaseQuantityLimit", "ValidatePurchaseQuantity",
	})
	assertFileDeclaresFunctions(t, filepath.Join(domainRoot, "inventory_policy.go"), []string{
		"ManualSKUAvailable", "ShouldEnforceManualSKUStock",
	})
	assertFileDeclaresFunctions(t, filepath.Join(manualFormRoot, "schema.go"), []string{
		"ValidateAndNormalize", "NormalizeSchema",
	})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "product_store.go"), []string{"ProductStore"})
	assertFileDeclaresFunctions(t, filepath.Join(storeRoot, "product_store.go"), []string{"NewProductStore"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "sku_store.go"), []string{"SKUStore"})
	assertFileDeclaresFunctions(t, filepath.Join(storeRoot, "sku_store.go"), []string{"NewSKUStore"})

	assertDirectoryGoFileBudget(t, productRoot, 3)
	assertDirectoryGoFileBudget(t, applicationRoot, 4)
	assertDirectoryGoFileBudget(t, adminRoot, 4)
	assertDirectoryGoFileBudget(t, writeRoot, 6)
	assertDirectoryGoFileBudget(t, domainRoot, 6)
	assertDirectoryGoFileBudget(t, manualFormRoot, 2)
	assertDirectoryGoFileBudget(t, storeRoot, 4)
	assertDirectoryGoFileBudget(t, sharedGORMRoot, 2)
}

func TestCatalogMappingImplementationLivesInBoundedContext(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	mappingRoot := filepath.Join(repositoryRoot, "internal", "modules", "catalog", "mapping")

	assertFileDeclaresTypes(t, filepath.Join(mappingRoot, "service.go"), []string{
		"ListFilter", "MappingRepository", "SKUMappingRepository", "ProductRepository", "SKURepository",
		"CategoryRepository", "ConnectionProvider", "MediaRecorder", "CategoryCreator", "SettingsProvider",
		"ImportTxProductRepository", "ImportTxSKURepository", "ImportTxMappingRepository", "ImportTxSKUMappingRepository",
		"ImportRepositories", "UnitOfWork", "ErrorSet", "Options", "Service",
	})
	assertFileDeclaresFunctions(t, filepath.Join(mappingRoot, "service.go"), []string{
		"NewService", "SetCategoryCreator", "SetSettings",
		"GetByID", "List", "SetActive", "Delete", "GetSKUMappings", "GetMappedUpstreamIDs",
	})
	assertFileDeclaresFunctions(t, filepath.Join(mappingRoot, "import.go"), []string{
		"ImportUpstreamProduct", "ImportUpstreamProductWithAutoCategory", "importUpstreamProduct",
		"createSKUMappings", "ListUpstreamProducts", "ListUpstreamCategories",
	})
	assertFileDeclaresTypes(t, filepath.Join(mappingRoot, "batch_import.go"), []string{
		"BatchUpstreamProductImportOutcome", "BatchImportByCategoryResult",
	})
	assertFileDeclaresFunctions(t, filepath.Join(mappingRoot, "batch_import.go"), []string{
		"BatchImportUpstreamProducts", "BatchImportByCategory",
		"findOrCreateCategoryFromUpstream", "findOrCreateLocalCategory",
	})
	assertFileDeclaresFunctions(t, filepath.Join(mappingRoot, "sync.go"), []string{
		"SyncProduct", "SyncAllStock", "SyncConnectionStock", "EnsureUpstreamStockForOrder",
		"markUpstreamUnavailable", "computeFullSyncInterval",
	})
	assertFileDeclaresFunctions(t, filepath.Join(mappingRoot, "markup.go"), []string{
		"ReapplyMarkup", "recalcProductPrice",
	})
	assertFileDeclaresFunctions(t, filepath.Join(mappingRoot, "pricing.go"), []string{
		"CalculateLocalPrice", "CalculateMarkedUpPrice", "convertCurrency",
	})

	storeRoot := filepath.Join(mappingRoot, "store", "gormstore")
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "mapping_store.go"), []string{"MappingStore"})
	assertFileDeclaresFunctions(t, filepath.Join(storeRoot, "mapping_store.go"), []string{"NewMappingStore"})
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "sku_mapping_store.go"), []string{"SKUMappingStore"})
	assertFileDeclaresFunctions(t, filepath.Join(storeRoot, "sku_mapping_store.go"), []string{"NewSKUMappingStore"})

	assertDirectoryGoFileBudget(t, mappingRoot, 12)
	assertDirectoryGoFileBudget(t, storeRoot, 4)
}

func TestCatalogMappingLegacyRepositoryFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	legacyFiles := []string{
		"internal/repository/product_mapping_repository.go",
		"internal/repository/product_mapping_repository_test.go",
		"internal/repository/sku_mapping_repository.go",
		"internal/repository/sku_mapping_repository_test.go",
	}
	for _, relativePath := range legacyFiles {
		if matches, err := filepath.Glob(filepath.Join(repositoryRoot, relativePath)); err != nil {
			t.Fatalf("glob %s: %v", relativePath, err)
		} else if len(matches) != 0 {
			t.Errorf("legacy Catalog mapping repository file must stay removed: %v", matches)
		}
	}
}

func TestCatalogMappingLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	legacyFiles := []string{
		"internal/service/product_mapping_import.go",
		"internal/service/product_mapping_batch_import.go",
		"internal/service/product_mapping_sync.go",
		"internal/service/product_mapping_markup.go",
		"internal/service/price_markup.go",
		"internal/service/price_markup_test.go",
	}
	for _, relativePath := range legacyFiles {
		if matches, err := filepath.Glob(filepath.Join(repositoryRoot, relativePath)); err != nil {
			t.Fatalf("glob %s: %v", relativePath, err)
		} else if len(matches) != 0 {
			t.Errorf("legacy Catalog mapping flat file must stay removed: %v", matches)
		}
	}
}

func TestCatalogProductLegacyRepositoryFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	legacyFiles := []string{
		"internal/repository/product_repository.go",
		"internal/repository/product_repository_test.go",
		"internal/repository/product_sku_repository.go",
		"internal/repository/product_sku_repository_test.go",
	}
	for _, relativePath := range legacyFiles {
		if matches, err := filepath.Glob(filepath.Join(repositoryRoot, relativePath)); err != nil {
			t.Fatalf("glob %s: %v", relativePath, err)
		} else if len(matches) != 0 {
			t.Errorf("legacy Catalog Product repository file must stay removed: %v", matches)
		}
	}
}

func TestCatalogProductLegacyFlatFilesStayRemoved(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	legacyFiles := []string{
		"internal/service/product_create.go",
		"internal/service/product_admin.go",
		"internal/service/product_query.go",
		"internal/service/product_rules.go",
		"internal/service/product_sku.go",
		"internal/service/product_stock.go",
		"internal/service/product_update.go",
		"internal/service/product_wholesale.go",
		"internal/service/product_purchase_limit.go",
		"internal/service/sku_stock_policy.go",
		"internal/service/manual_form_validator.go",
		"internal/service/manual_form_validator_test.go",
		"internal/http/handlers/admin/admin_product.go",
		"internal/http/handlers/admin/admin_product_test.go",
		"internal/http/handlers/admin/admin_product_mapping.go",
		"internal/http/handlers/admin/admin_product_mapping_test.go",
	}
	for _, relativePath := range legacyFiles {
		if matches, err := filepath.Glob(filepath.Join(repositoryRoot, relativePath)); err != nil {
			t.Fatalf("glob %s: %v", relativePath, err)
		} else if len(matches) != 0 {
			t.Errorf("legacy Catalog Product flat file must stay removed: %v", matches)
		}
	}
}
