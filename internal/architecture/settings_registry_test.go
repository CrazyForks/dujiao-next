package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSettingServiceRoutesWritesThroughRegistry(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)

	// Core Update/UpdateWithEffects now live in modules/settings; the service facade must embed that module.
	facadePath := filepath.Join(repositoryRoot, "internal", "service", "setting_service.go")
	facadeParsed, err := parser.ParseFile(token.NewFileSet(), facadePath, nil, 0)
	if err != nil {
		t.Fatalf("parse setting service facade: %v", err)
	}

	importsModule := false
	for _, imported := range facadeParsed.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			t.Fatalf("unquote setting service import: %v", err)
		}
		if importPath == moduleImportPath+"/internal/modules/settings" {
			importsModule = true
		}
	}
	if !importsModule {
		t.Fatal("SettingService must import modules/settings")
	}

	embedsModuleService := false
	for _, declaration := range facadeParsed.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok || generic.Tok != token.TYPE {
			continue
		}
		for _, specification := range generic.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "SettingService" {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				if len(field.Names) != 0 {
					continue
				}
				star, ok := field.Type.(*ast.StarExpr)
				if !ok {
					continue
				}
				selector, ok := star.X.(*ast.SelectorExpr)
				if ok && selector.Sel.Name == "Service" {
					embedsModuleService = true
				}
			}
		}
	}
	if !embedsModuleService {
		t.Fatal("SettingService must embed *modules/settings.Service")
	}

	modulePath := filepath.Join(repositoryRoot, "internal", "modules", "settings", "service.go")
	moduleParsed, err := parser.ParseFile(token.NewFileSet(), modulePath, nil, 0)
	if err != nil {
		t.Fatalf("parse settings module service: %v", err)
	}

	updateFound := false
	updateDelegates := false
	updateWithEffectsFound := false
	usesRegistry := false
	for _, declaration := range moduleParsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		switch function.Name.Name {
		case "Update":
			updateFound = true
		case "UpdateWithEffects":
			updateWithEffectsFound = true
		default:
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			switch expression := node.(type) {
			case *ast.SwitchStmt, *ast.TypeSwitchStmt:
				t.Errorf("settings.Service.%s must delegate by registry instead of a key switch", function.Name.Name)
			case *ast.SelectorExpr:
				if function.Name.Name == "Update" && expression.Sel.Name == "UpdateWithEffects" {
					updateDelegates = true
				}
				if function.Name.Name == "UpdateWithEffects" && expression.Sel.Name == "Normalize" {
					usesRegistry = true
				}
			}
			return true
		})
	}
	if !updateFound {
		t.Fatal("settings.Service.Update not found")
	}
	if !updateDelegates {
		t.Fatal("settings.Service.Update must preserve compatibility by delegating to UpdateWithEffects")
	}
	if !updateWithEffectsFound || !usesRegistry {
		t.Fatal("settings.Service.UpdateWithEffects must normalize writes through Registry")
	}
}

func TestLegacySettingKeyDispatcherIsAbsent(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceFiles, err := filepath.Glob(filepath.Join(repositoryRoot, "internal", "service", "*.go"))
	if err != nil {
		t.Fatalf("list service files: %v", err)
	}

	for _, serviceFile := range serviceFiles {
		if filepath.Ext(serviceFile) != ".go" {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), serviceFile, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", serviceFile, err)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Name.Name == "normalizeSettingValueByKey" {
				t.Errorf("%s reintroduces the legacy global setting key dispatcher", filepath.Base(serviceFile))
			}
		}
	}
}

func TestAdminSettingHandlerDispatchesRegistryEffectsWithoutKeyComparisons(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	handlerPath := filepath.Join(repositoryRoot, "internal", "transport", "http", "settings", "admin_handler.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), handlerPath, nil, 0)
	if err != nil {
		t.Fatalf("parse admin setting handler: %v", err)
	}

	updateWithEffects := false
	hasEffect := false
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "Update" || function.Body == nil {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			switch expression := node.(type) {
			case *ast.SelectorExpr:
				switch expression.Sel.Name {
				case "UpdateWithEffects":
					updateWithEffects = true
				case "HasEffect":
					hasEffect = true
				}
			case *ast.BinaryExpr:
				if selector, ok := expression.X.(*ast.SelectorExpr); ok && selector.Sel.Name == "Key" {
					if identifier, ok := selector.X.(*ast.Ident); ok && identifier.Name == "req" {
						t.Errorf("Update must dispatch registry effects instead of comparing req.Key")
					}
				}
			}
			return true
		})
	}
	if !updateWithEffects || !hasEffect {
		t.Fatalf("Update must consume UpdateWithEffects and HasEffect; update=%v hasEffect=%v", updateWithEffects, hasEffect)
	}
}

func TestSettingRegistryUsesModuleOwnedTypedJSONNormalizers(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	path := filepath.Join(repositoryRoot, "internal", "service", "setting_registry.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse setting registry: %v", err)
	}

	want := map[string]bool{
		"NormalizeDashboardSettingJSON":       false,
		"NormalizeAffiliateSettingJSON":       false,
		"NormalizeUpstreamSyncConfigJSON":     false,
		"NormalizeTelegramAuthSettingJSON":    false,
		"NormalizeTelegramBotConfigJSON":      false,
		"NormalizeCallbackRoutesSettingJSON":  false,
		"NormalizeOrderRiskControlConfigJSON": false,
		"NormalizeHomeAnnouncementJSON":       false,
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if _, expected := want[selector.Sel.Name]; expected {
			want[selector.Sel.Name] = true
		}
		return true
	})
	for normalizer, found := range want {
		if !found {
			t.Errorf("setting registry must use module-owned %s", normalizer)
		}
	}
}
