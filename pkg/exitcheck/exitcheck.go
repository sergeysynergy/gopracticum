// Package exitcheck Определяет анализатор на базе analysis.Analyzer,
// который отслеживает прямой вызов функции os.Exit.
package exitcheck

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name:             "exitcheck",
	Doc:              "check for direct os.Exit call in main function of main package",
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	Run:              run,
	RunDespiteErrors: true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			exitCheck(file, pass)
		}
	}
	return nil, nil
}

func exitCheck(file *ast.File, pass *analysis.Pass) {
	ast.Inspect(file, func(node ast.Node) bool {
		for _, v := range file.Decls {
			if mainFnc, ok := v.(*ast.FuncDecl); ok {
				if mainFnc.Name.Name != "main" {
					continue
				}
				// Рекурсивно проходим по элементам узла функция main: ищем прямое объявление os.Exit().
				ast.Inspect(mainFnc, func(node ast.Node) bool {
					switch x := node.(type) {
					case *ast.SelectorExpr:
						if fmt.Sprint(x.X) == "os" && x.Sel.Name == "Exit" {
							pass.Report(analysis.Diagnostic{
								Pos:     x.Pos(),
								Message: "direct function call `os.Exit()` in main package",
							})
						}
					}
					return true
				})
			}
		}
		return true
	})
}
