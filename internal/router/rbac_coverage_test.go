package router

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/authz"

	"github.com/casbin/casbin/v3/util"
)

// TestAllAdminRoutesCoveredByBuiltinRoles 校验 admin 路由文件里的每条路由
// 都被 authz.BuiltinRoleSeeds() 中至少一条角色策略覆盖。
//
// 目的：避免新增 admin 接口时忘记同步 RBAC 预置角色，导致非超管角色无法通过
// 角色分配获得该权限（catalog UI 上能看到，但任何角色都拿不到）。
//
// 实现：静态扫描 routes_admin*.go 提取 authorized.METHOD("/path", ...) 与
// paymentProtected.METHOD("/path", ...) 调用，与
// builtin role seeds 用 keyMatch2 比对（与运行时 Casbin 模型一致）。
func TestAllAdminRoutesCoveredByBuiltinRoles(t *testing.T) {
	routes, err := extractAdminRoutesFromSource()
	if err != nil {
		t.Fatalf("extract admin routes: %v", err)
	}
	if len(routes) == 0 {
		t.Fatalf("no admin routes extracted; regex or source layout changed?")
	}

	seeds := authz.BuiltinRoleSeeds()
	if len(seeds) == 0 {
		t.Fatalf("no builtin role seeds")
	}

	type policy struct {
		object string
		action string
	}
	var policies []policy
	for _, seed := range seeds {
		for _, p := range seed.Policies {
			policies = append(policies, policy{
				object: authz.NormalizeObject(p.Object),
				action: authz.NormalizeAction(p.Action),
			})
		}
	}

	var uncovered []adminRoute
	for _, r := range routes {
		matched := false
		for _, p := range policies {
			if p.action != "*" && p.action != r.method {
				continue
			}
			if util.KeyMatch2(r.object, p.object) {
				matched = true
				break
			}
		}
		if !matched {
			uncovered = append(uncovered, r)
		}
	}

	if len(uncovered) > 0 {
		var lines []string
		for _, r := range uncovered {
			lines = append(lines, "  "+r.method+" "+r.object)
		}
		t.Fatalf("the following admin routes are not covered by any builtin role seed in authz.BuiltinRoleSeeds() — add them to the appropriate role in api/internal/authz/bootstrap.go:\n%s",
			strings.Join(lines, "\n"))
	}
}

type adminRoute struct {
	method string
	object string // 例如 "/admin/users/:id"
}

// extractAdminRoutesFromSource 从 admin 路由文件和已抽取的模块路由文件中读取调用。
// 方法范围：GET / POST / PUT / PATCH / DELETE。HEAD/OPTIONS 不参与 RBAC。
func extractAdminRoutesFromSource() ([]adminRoute, error) {
	_, thisFile, _, _ := runtime.Caller(0)
	routerDirectory := filepath.Dir(thisFile)
	type routeSource struct {
		path       string
		expression *regexp.Regexp
	}

	adminSources, err := filepath.Glob(filepath.Join(routerDirectory, "routes_admin*.go"))
	if err != nil {
		return nil, err
	}
	if len(adminSources) == 0 {
		return nil, os.ErrNotExist
	}

	sources := make([]routeSource, 0, len(adminSources)+8)
	for _, path := range adminSources {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		sources = append(sources, routeSource{
			path:       path,
			expression: regexp.MustCompile(`(?:authorized|paymentProtected)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		})
	}
	if len(sources) == 0 {
		return nil, os.ErrNotExist
	}
	sources = append(sources,
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "transport", "http", "content", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "transport", "http", "dashboard", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "modules", "memberlevel", "transport", "http", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "modules", "apicredential", "transport", "http", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "modules", "auditlog", "transport", "http", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "modules", "coupon", "transport", "http", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "modules", "promotion", "transport", "http", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "modules", "giftcard", "transport", "http", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
		routeSource{
			path:       filepath.Join(routerDirectory, "..", "transport", "http", "notification", "routes.go"),
			expression: regexp.MustCompile(`admin\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`),
		},
	)

	seen := make(map[string]struct{})
	var out []adminRoute
	for _, source := range sources {
		raw, err := os.ReadFile(source.path)
		if err != nil {
			return nil, err
		}
		for _, match := range source.expression.FindAllStringSubmatch(string(raw), -1) {
			method := match[1]
			object := authz.NormalizeObject("/admin" + match[2])
			key := method + " " + object
			if _, duplicate := seen[key]; duplicate {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, adminRoute{method: method, object: object})
		}
	}
	return out, nil
}
