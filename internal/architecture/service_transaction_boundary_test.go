package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceTransactionsUseRepositoryMethods(t *testing.T) {
	repositoryRoot := findRepositoryRoot(t)
	serviceDirectory := filepath.Join(repositoryRoot, "internal", "service")
	entries, err := os.ReadDir(serviceDirectory)
	if err != nil {
		t.Fatalf("read service package: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(serviceDirectory, entry.Name())
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			receiver, ok := selector.X.(*ast.Ident)
			if !ok || receiver.Name != "tx" {
				return true
			}

			position := fileSet.Position(call.Pos())
			relativePath, relErr := filepath.Rel(repositoryRoot, position.Filename)
			if relErr != nil {
				relativePath = position.Filename
			}
			t.Errorf(
				"%s:%d calls tx.%s directly; keep transaction callbacks in the service layer and execute database operations through repository WithTx methods",
				filepath.ToSlash(relativePath),
				position.Line,
				selector.Sel.Name,
			)
			return true
		})
	}
}
