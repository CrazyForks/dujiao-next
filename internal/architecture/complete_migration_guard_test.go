package architecture

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// legacyRootGoFileBudgets freeze the remaining horizontal architecture at the
// start of the complete migration. Every entry is a one-way budget: deleting
// or moving files is allowed, adding files below these roots is not. The map is
// removed together with the roots when the migration is complete.
var legacyRootGoFileBudgets = map[string]int{
	"internal/dto":            24,
	"internal/http":           17,
	"internal/integration":    25,
	"internal/models":         59,
	"internal/payment":        53,
	"internal/persistence":    3,
	"internal/provider":       10,
	"internal/repository":     54,
	"internal/router":         23,
	"internal/service":        141,
	"internal/transport/http": 184,
	"internal/wiring":         51,
	"internal/worker":         10,
}

// These are the only production compatibility shims that existed at the
// migration baseline. They may be deleted, but no replacement shim may be
// introduced. The final migration gate requires this set to become empty.
var baselineCompatibilityFiles = map[string]struct{}{
	"internal/repository/card_secret_compat.go": {},
}

type packageFileBudget struct {
	production int
	total      int
}

// Existing oversized packages are frozen at their baseline size. All new or
// already-focused packages use the default limit below. Entries disappear as
// their packages are split into bounded-context leaf packages.
var transitionalPackageFileBudgets = map[string]packageFileBudget{
	"internal/architecture":            {production: 0, total: 54},
	"internal/dto":                     {production: 15, total: 24},
	"internal/models":                  {production: 54, total: 59},
	"internal/modules/reseller":        {production: 19, total: 26},
	"internal/payment/provider":        {production: 14, total: 25},
	"internal/repository":              {production: 36, total: 54},
	"internal/router":                  {production: 13, total: 23},
	"internal/service":                 {production: 81, total: 141},
	"internal/transport/http/reseller": {production: 13, total: 20},
}

// completedMigrationPaths are deleted compatibility-free entry points. Once a
// bounded context reaches this list, recreating its former horizontal package
// is an architecture regression rather than an allowed transitional change.
var completedMigrationPaths = []string{
	"internal/http/handlers/shared/captcha_payload.go",
	"internal/models/setting.go",
	"internal/repository/setting_repository.go",
	"internal/service/setting_service.go",
	"internal/transport/http/captcha",
	"internal/transport/http/settings",
	"internal/wiring/captcha",
	"internal/wiring/settings",
	"internal/models/telegram_broadcast.go",
	"internal/repository/telegram_broadcast_repository.go",
	"internal/service/telegram_broadcast_service.go",
	"internal/transport/http/telegram/admin_broadcast_handler.go",
	"internal/wiring/telegram/broadcast.go",
	"internal/models/channel_client.go",
	"internal/repository/channel_client_repository.go",
	"internal/transport/http/channelclient",
	"internal/wiring/channelclient",
	"internal/wiring/telegram/channel_bot.go",
	"internal/service/telegram_auth_service.go",
	"internal/service/telegram_auth_service_test.go",
	"internal/service/telegram_oidc.go",
	"internal/service/telegram_oidc_test.go",
	"internal/models/email_verify_code.go",
	"internal/repository/email_verify_code_repository.go",
	"internal/models/user_oauth_identity.go",
	"internal/repository/user_oauth_identity_repository.go",
	"internal/repository/user_oauth_identity_repository_test.go",
	"internal/models/admin.go",
	"internal/models/init.go",
	"internal/models/init_test.go",
	"internal/repository/admin_repository.go",
	"internal/repository/admin_repository_test.go",
	"internal/models/money.go",
	"internal/models/user.go",
	"internal/repository/user_repository.go",
	"internal/repository/user_repository_test.go",
	"internal/service/totp_enable.go",
	"internal/service/totp_enable_test.go",
	"internal/service/recovery_codes.go",
	"internal/service/password_policy.go",
	"internal/service/jwt_parser.go",
	"internal/service/user_auth_service.go",
	"internal/service/user_auth_service_profile.go",
	"internal/service/user_auth_telegram_binding.go",
	"internal/service/user_auth_telegram_channel.go",
	"internal/service/user_auth_telegram_identity.go",
	"internal/service/user_auth_telegram_login.go",
	"internal/service/user_auth_telegram_oidc.go",
	"internal/service/user_auth_service_2fa_test.go",
	"internal/service/user_auth_service_channel_identity_test.go",
	"internal/service/user_auth_service_domain_policy_test.go",
	"internal/service/user_auth_service_email_mode_test.go",
	"internal/service/user_auth_service_oauth_miniapp_test.go",
	"internal/service/user_auth_service_oauth_test.go",
	"internal/service/user_auth_telegram_test_helpers_test.go",
	"internal/service/user_totp_service.go",
	"internal/service/user_totp_service_test.go",
	"internal/service/auth_service.go",
	"internal/service/auth_service_test.go",
	"internal/service/totp_service.go",
	"internal/service/totp_service_test.go",
	"internal/models/affiliate_profile.go",
	"internal/models/affiliate_click.go",
	"internal/models/affiliate_commission.go",
	"internal/models/affiliate_withdraw_request.go",
	"internal/repository/affiliate_repository.go",
	"internal/service/affiliate_service.go",
	"internal/service/affiliate_attribution.go",
	"internal/service/affiliate_commission.go",
	"internal/service/affiliate_profile.go",
	"internal/service/affiliate_query.go",
	"internal/service/affiliate_withdraw.go",
	"internal/service/affiliate_service_test.go",
	"internal/transport/http/affiliate",
	"internal/wiring/affiliate",
	"internal/modules/affiliate/errors.go",
	"internal/modules/affiliate/types.go",
	"internal/dto/affiliate.go",
	"internal/dto/affiliate_test.go",
	"internal/integration/affiliate",
	"internal/service/product_service.go",
	"internal/service/product_application_compat.go",
	"internal/repository/product_compat.go",
	"internal/repository/product_mapping_compat.go",
	"internal/service/product_mapping_service.go",
	"internal/service/product_mapping_service_test.go",
	"internal/models/category.go",
	"internal/modules/catalog/category_service.go",
	"internal/modules/catalog/store/gormstore/category_store.go",
	"internal/modules/catalog/store/gormstore/category_store_test.go",
	"internal/transport/http/catalog/admin_category_handler.go",
	"internal/integration/catalog/category_service_test.go",
	"internal/dto/product_test.go",
	"internal/models/product.go",
	"internal/models/product_sku.go",
	"internal/models/wholesale_price.go",
	"internal/modules/catalog/product/errors.go",
	"internal/modules/catalog/product/ports.go",
	"internal/dto/product.go",
	"internal/transport/http/catalog/admin_product_handler.go",
	"internal/transport/http/catalog/admin_product_handler_integration_test.go",
	"internal/transport/http/catalog/public_handler.go",
	"internal/transport/http/catalog/public_view.go",
	"internal/transport/http/catalog/public_stock.go",
	"internal/transport/http/catalog/public_price_integration_test.go",
	"internal/transport/http/catalog/public_price_test.go",
	"internal/transport/http/catalog/public_related_posts_test.go",
	"internal/transport/http/catalog/public_stock_test.go",
	"internal/transport/http/catalog/admin_product_mapping_handler.go",
	"internal/transport/http/catalog/admin_product_mapping_handler_integration_test.go",
	"internal/transport/http/catalog/routes.go",
	"internal/transport/http/catalog",
	"internal/wiring/catalog/factory.go",
	"internal/wiring/catalog/wiring.go",
	"internal/wiring/catalog",
	"internal/service/product_admin_test.go",
	"internal/service/product_create_test.go",
	"internal/service/product_query_test.go",
	"internal/service/product_reseller_public_test.go",
	"internal/service/product_sku_test.go",
	"internal/service/product_stock_test.go",
	"internal/service/product_test_helpers_test.go",
	"internal/service/product_update_test.go",
	"internal/service/product_wholesale_test.go",
	"internal/models/site_connection.go",
	"internal/service/site_connection_service.go",
	"internal/http/handlers/admin/admin_site_connection.go",
	"internal/modules/siteconnection/errors.go",
	"internal/modules/siteconnection/service.go",
	"internal/modules/siteconnection/types.go",
	"internal/repository/site_connection_repository.go",
	"internal/transport/http/siteconnection",
	"internal/integration/siteconnection",
	"internal/wiring/siteconnection",
	"internal/models/product_mapping.go",
	"internal/modules/catalog/mapping/batch_import.go",
	"internal/modules/catalog/mapping/batch_import_test.go",
	"internal/modules/catalog/mapping/import.go",
	"internal/modules/catalog/mapping/import_test.go",
	"internal/modules/catalog/mapping/markup.go",
	"internal/modules/catalog/mapping/pricing.go",
	"internal/modules/catalog/mapping/pricing_test.go",
	"internal/modules/catalog/mapping/service.go",
	"internal/modules/catalog/mapping/service_test.go",
	"internal/modules/catalog/mapping/sync.go",
	"internal/modules/catalog/mapping/sync_test.go",
	"internal/modules/catalog/mapping/store",
	"internal/models/member_level.go",
	"internal/models/member_level_price.go",
	"internal/modules/memberlevel/ports.go",
	"internal/modules/memberlevel/service.go",
	"internal/modules/memberlevel/store",
	"internal/transport/http/memberlevel",
	"internal/models/promotion.go",
	"internal/modules/promotion/admin_service.go",
	"internal/modules/promotion/errors.go",
	"internal/modules/promotion/ports.go",
	"internal/modules/promotion/service.go",
	"internal/modules/promotion/store",
	"internal/transport/http/promotion",
	"internal/models/coupon.go",
	"internal/models/coupon_usage.go",
	"internal/modules/coupon/admin_service.go",
	"internal/modules/coupon/errors.go",
	"internal/modules/coupon/ports.go",
	"internal/modules/coupon/service.go",
	"internal/modules/coupon/store",
	"internal/transport/http/coupon",
	"internal/models/gift_card.go",
	"internal/models/gift_card_batch.go",
	"internal/modules/giftcard/errors.go",
	"internal/modules/giftcard/export.go",
	"internal/modules/giftcard/generate.go",
	"internal/modules/giftcard/helpers.go",
	"internal/modules/giftcard/helpers_test.go",
	"internal/modules/giftcard/manage.go",
	"internal/modules/giftcard/ports.go",
	"internal/modules/giftcard/service.go",
	"internal/modules/giftcard/types.go",
	"internal/modules/giftcard/store",
	"internal/service/gift_card_service.go",
	"internal/service/gift_card_service_test.go",
	"internal/service/gift_card_store_adapter.go",
	"internal/dto/gift_card.go",
	"internal/transport/http/giftcard",
	"internal/integration/channel/giftcard_test.go",
	"internal/models/api_credential.go",
	"internal/modules/apicredential/ports.go",
	"internal/modules/apicredential/service.go",
	"internal/modules/apicredential/store",
	"internal/transport/http/apicredential",
	"internal/models/admin_login_log.go",
	"internal/models/user_login_log.go",
	"internal/models/authz_audit_log.go",
	"internal/modules/auditlog/authz_service.go",
	"internal/modules/auditlog/ports.go",
	"internal/modules/auditlog/service_test.go",
	"internal/modules/auditlog/store",
	"internal/modules/auditlog/user_login_service.go",
	"internal/transport/http/auditlog",
	"internal/dto/login_log.go",
	"internal/models/notification_log.go",
	"internal/modules/notification/alert.go",
	"internal/modules/notification/dedupe.go",
	"internal/modules/notification/errors.go",
	"internal/modules/notification/inventory_alert.go",
	"internal/modules/notification/log_service.go",
	"internal/modules/notification/log_service_test.go",
	"internal/modules/notification/order_format.go",
	"internal/modules/notification/payment_order_alert.go",
	"internal/modules/notification/ports.go",
	"internal/modules/notification/send.go",
	"internal/modules/notification/service.go",
	"internal/modules/notification/store",
	"internal/modules/notification/template.go",
	"internal/modules/notification/test_variables.go",
	"internal/transport/http/notification",
	"internal/models/post.go",
	"internal/models/post_category.go",
	"internal/models/banner.go",
	"internal/models/media.go",
	"internal/dto/post.go",
	"internal/dto/banner.go",
	"internal/dto/banner_test.go",
	"internal/modules/content/banner_service.go",
	"internal/modules/content/errors.go",
	"internal/modules/content/media_service.go",
	"internal/modules/content/ports.go",
	"internal/modules/content/post_category_service.go",
	"internal/modules/content/post_service.go",
	"internal/modules/content/query.go",
	"internal/modules/content/service_test.go",
	"internal/modules/content/system.go",
	"internal/modules/content/store",
	"internal/modules/content/filestore",
	"internal/transport/http/content",
}

func TestCompletedMigrationPathsStayDeleted(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	for _, relativePath := range completedMigrationPaths {
		t.Run(strings.ReplaceAll(relativePath, "/", "_"), func(t *testing.T) {
			absolutePath := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
			if _, err := os.Stat(absolutePath); err == nil {
				t.Fatalf("completed migration path was recreated: %s", relativePath)
			} else if !os.IsNotExist(err) {
				t.Fatalf("stat completed migration path %s: %v", relativePath, err)
			}
		})
	}
}

func TestSettingsModuleRootContainsNoGoFiles(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	settingsRoot := filepath.Join(repositoryRoot, "internal", "modules", "settings")
	production, total := countDirectGoFiles(t, settingsRoot)
	if production != 0 || total != 0 {
		t.Fatalf("settings module root must remain structural only, got production=%d total=%d", production, total)
	}
}

func TestSharedMoneyOwnsMonetaryValueObject(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	moneyRoot := filepath.Join(repositoryRoot, "internal", "shared", "money")
	assertFileDeclaresTypes(t, filepath.Join(moneyRoot, "amount.go"), []string{"Amount"})
	assertFileDeclaresFunctions(t, filepath.Join(moneyRoot, "amount.go"), []string{"FromDecimal"})
	assertDirectoryGoFileBudget(t, moneyRoot, 2)
}

func TestSharedPasswordPolicyOwnsStrengthValidation(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	policyRoot := filepath.Join(repositoryRoot, "internal", "shared", "passwordpolicy")
	assertFileDeclaresTypes(t, filepath.Join(policyRoot, "policy.go"), []string{"Policy"})
	assertFileDeclaresFunctions(t, filepath.Join(policyRoot, "policy.go"), []string{"Validate"})
	assertDirectoryGoFileBudget(t, policyRoot, 2)
}

func TestAffiliateModuleOwnsDomainAndPersistence(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	domainRoot := filepath.Join(repositoryRoot, "internal", "modules", "affiliate", "domain")
	assertFileDeclaresTypes(t, filepath.Join(domainRoot, "profile.go"), []string{"Profile"})
	assertFileDeclaresTypes(t, filepath.Join(domainRoot, "click.go"), []string{"Click"})
	assertFileDeclaresTypes(t, filepath.Join(domainRoot, "commission.go"), []string{"Commission"})
	assertFileDeclaresTypes(t, filepath.Join(domainRoot, "withdraw_request.go"), []string{"WithdrawRequest"})
	assertDirectoryGoFileBudget(t, domainRoot, 5)

	storeRoot := filepath.Join(repositoryRoot, "internal", "modules", "affiliate", "infrastructure", "gormstore")
	assertFileDeclaresTypes(t, filepath.Join(storeRoot, "store.go"), []string{"Store"})
	assertDirectoryGoFileBudget(t, storeRoot, 2)

	moduleRoot := filepath.Join(repositoryRoot, "internal", "modules", "affiliate")
	production, total := countDirectGoFiles(t, moduleRoot)
	if production != 0 || total != 0 {
		t.Fatalf("affiliate module root must remain structural only, got production=%d total=%d", production, total)
	}
}

func TestLegacyHorizontalRootsCanOnlyShrink(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)

	paths := make([]string, 0, len(legacyRootGoFileBudgets))
	for relativePath := range legacyRootGoFileBudgets {
		paths = append(paths, relativePath)
	}
	sort.Strings(paths)

	for _, relativePath := range paths {
		t.Run(strings.TrimPrefix(relativePath, "internal/"), func(t *testing.T) {
			absolutePath := filepath.Join(repositoryRoot, filepath.FromSlash(relativePath))
			count := countGoFilesRecursively(t, absolutePath)
			if count > legacyRootGoFileBudgets[relativePath] {
				t.Fatalf(
					"legacy root %s grew from its migration baseline of %d Go files to %d; move new code into a bounded context",
					relativePath,
					legacyRootGoFileBudgets[relativePath],
					count,
				)
			}
		})
	}
}

func TestNoNewCompatibilityOrLegacyProductionFiles(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	internalRoot := filepath.Join(repositoryRoot, "internal")

	err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		name := strings.ToLower(entry.Name())
		if !strings.Contains(name, "compat") && !strings.Contains(name, "legacy") && !strings.Contains(name, "facade") {
			return nil
		}

		relativePath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)
		if _, allowedDuringMigration := baselineCompatibilityFiles[relativePath]; !allowedDuringMigration {
			t.Errorf("new compatibility or legacy production file is forbidden: %s", relativePath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect production Go files: %v", err)
	}
}

func TestGoPackagesStayWithinFileBudgets(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	internalRoot := filepath.Join(repositoryRoot, "internal")

	err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() || entry.Name() == "testdata" {
			return nil
		}

		production, total := countDirectGoFiles(t, path)
		if total == 0 {
			return nil
		}

		relativePath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)

		budget := packageFileBudget{production: 12, total: 20}
		if transitional, ok := transitionalPackageFileBudgets[relativePath]; ok {
			budget = transitional
		}

		if production > budget.production || total > budget.total {
			t.Errorf(
				"Go package %s exceeds its migration budget: production=%d/%d total=%d/%d",
				relativePath,
				production,
				budget.production,
				total,
				budget.total,
			)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect Go package budgets: %v", err)
	}
}

func countGoFilesRecursively(t *testing.T, root string) int {
	t.Helper()
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("stat %s: %v", root, err)
	}

	count := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && filepath.Ext(path) == ".go" {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("count Go files below %s: %v", root, err)
	}
	return count
}

func countDirectGoFiles(t *testing.T, directory string) (production int, total int) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read %s: %v", directory, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		total++
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			production++
		}
	}
	return production, total
}
