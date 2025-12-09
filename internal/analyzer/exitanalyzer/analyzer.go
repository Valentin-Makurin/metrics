package exitanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = `exitanalyzer проверяет отсутствие прямого вызова os.Exit в функции main пакета main`

var Analyzer = &analysis.Analyzer{
	Name: "exitanalyzer",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		if funcDecl.Name.Name != "main" {
			return
		}

		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := selectorExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			// Проверяем вызов os.Exit
			if ident.Name == "os" && selectorExpr.Sel.Name == "Exit" {
				pass.Reportf(callExpr.Pos(),
					"прямой вызов os.Exit в функции main запрещен. ")
			}

			return true
		})
	})

	return nil, nil
}
