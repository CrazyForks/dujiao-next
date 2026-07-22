package architecture

import (
	"go/ast"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestProductServiceImplementationIsSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	expected := map[string][]string{
		"product_service.go": {"NewProductService"},
		"product_application_compat.go": {
			"newProductWriteUnitOfWork", "WithinTransaction",
			"newProductAdminUnitOfWork", "WithinTransaction", "DeleteByProduct",
			"bindMappingDeleteTx",
		},
		"product_mapping_service.go": {
			"NewProductMappingService", "SetCategoryService", "SetSettingService",
			"newProductMappingUnitOfWork", "WithinTransaction",
			"bindMappingImportTx", "bindSKUMappingImportTx",
		},
	}

	for file, want := range expected {
		path := filepath.Join(serviceDirectory, file)
		parsed := parseProductionGoFile(t, path)
		got := declaredFunctionNames(parsed)
		sort.Strings(want)
		sort.Strings(got)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s function ownership mismatch\nwant: %v\ngot:  %v", file, want, got)
		}
	}

	base := parseProductionGoFile(t, filepath.Join(serviceDirectory, "product_service.go"))
	for _, declaration := range base.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv != nil {
			t.Errorf("product_service.go must only declare ProductService dependencies and DTOs; method %s belongs in a responsibility file", function.Name.Name)
		}
	}
}

func TestProductServiceTestsAreSplitByResponsibility(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	legacyPath := filepath.Join(serviceDirectory, "product_service_test.go")
	if _, err := os.Stat(legacyPath); err == nil {
		t.Fatalf("product_service_test.go must be replaced by responsibility-focused test files")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat product_service_test.go: %v", err)
	}

	expectedOwner := map[string]string{
		"TestProductServiceUpdateRejectsDisablingAutoSKUWithCardSecretStock": "product_sku_test.go",
		"TestApplyAutoStockCounts_LegacyStockPrefersDefaultSKU":              "product_stock_test.go",
		"TestProductServiceListPublicIncludesChildProductsForParentCategory": "product_query_test.go",
		"TestProductServiceListPublicSortOrderDescending":                    "product_query_test.go",
		"TestProductServiceListPublicSortsSKUsDescending":                    "product_query_test.go",
		"TestProductServiceGetAdminByIDIncludesInactiveSKUs":                 "product_query_test.go",
		"TestProductServiceCreateRejectsParentCategoryWithChildren":          "product_create_test.go",
		"TestProductServiceCreateFiltersUnavailablePaymentChannels":          "product_create_test.go",
		"TestProductServiceCreateRejectsInvalidPurchaseLimits":               "product_create_test.go",
		"TestProductServiceUpdateKeepsMappedProductFulfillmentUpstream":      "product_update_test.go",
		"TestProductServiceUpdateFiltersUnavailablePaymentChannels":          "product_update_test.go",
		"TestProductServiceUpdateRejectsInvalidPurchaseLimits":               "product_update_test.go",
		"TestProductServiceQuickUpdateRejectsActivationWithoutCategory":      "product_admin_test.go",
		"TestProductServiceDeleteCascade":                                    "product_admin_test.go",
		"TestProductServiceDeleteRollsBackCascadeWhenProductDeleteFails":     "product_admin_test.go",
		"TestProductServiceUpdateWholesalePricesOptionalSemantics":           "product_wholesale_test.go",
		"TestProductServiceUpdateWholesalePricesOnlyTouchesWholesaleField":   "product_wholesale_test.go",
		"TestProductServiceUpdateWholesalePricesClearsTiers":                 "product_wholesale_test.go",
		"TestProductServiceUpdateWholesalePricesRejectsInvalidInputs":        "product_wholesale_test.go",
		"TestProductServiceUpdateWholesalePricesValidatesSKUBelonging":       "product_wholesale_test.go",
		"TestProductServiceUpdateWholesalePricesReturnsNotFound":             "product_wholesale_test.go",
	}

	actualOwners := make(map[string][]string, len(expectedOwner))
	for _, file := range []string{
		"product_sku_test.go",
		"product_stock_test.go",
		"product_query_test.go",
		"product_create_test.go",
		"product_update_test.go",
		"product_admin_test.go",
		"product_wholesale_test.go",
	} {
		parsed := parseProductionGoFile(t, filepath.Join(serviceDirectory, file))
		for _, function := range declaredFunctionNames(parsed) {
			if _, tracked := expectedOwner[function]; tracked {
				actualOwners[function] = append(actualOwners[function], file)
			}
		}
	}

	for function, wantFile := range expectedOwner {
		gotFiles := actualOwners[function]
		if len(gotFiles) != 1 || gotFiles[0] != wantFile {
			t.Errorf("%s ownership mismatch: want [%s], got %v", function, wantFile, gotFiles)
		}
	}
}

func declaredFunctionNames(parsed *ast.File) []string {
	functions := make([]string, 0)
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		functions = append(functions, function.Name.Name)
	}
	return functions
}
