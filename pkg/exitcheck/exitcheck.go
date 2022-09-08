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
	var err error
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			err = exitCheck(file, pass)
		}
	}
	return nil, err
}

func exitCheck(file *ast.File, pass *analysis.Pass) error {
	var err error

	ast.Inspect(file, func(node ast.Node) bool {
		for _, v := range file.Decls {
			if mainFnc, ok := v.(*ast.FuncDecl); ok {
				if mainFnc.Name.Name != "main" {
					continue
				}
				// Рекурсивно проходим по элементам узла функция main: ищем прямое объявление os.Exit().
				ast.Inspect(mainFnc, func(node ast.Node) bool {
					switch x := node.(type) {
					/* Вариант 1: ищем просто вызов os.Exit()
					case *ast.SelectorExpr:
						if fmt.Sprint(x.X) == "os" && x.Sel.Name == "Exit" {
							pass.Report(analysis.Diagnostic{
								Pos:     x.Pos(),
								Message: "direct function call `os.Exit()` in main package",
							})
						}
					*/
					// Вариант 2: ищем присваивания функции os.Exit()
					case *ast.AssignStmt:
						for i := 0; i < len(x.Lhs); i++ {
							if _, okAss := x.Lhs[i].(*ast.Ident); !okAss {
								continue
							}
							// вызов функции справа
							call, okAss := x.Rhs[i].(*ast.CallExpr)
							if !okAss {
								continue
							}
							fn, okCall := call.Fun.(*ast.SelectorExpr)
							if !okCall {
								continue
							}
							// непосредственная проверка элементов дерева на соответствие названию os.Exit
							if fmt.Sprint(fn.X) == "os" && fn.Sel.Name == "Exit" {
								pass.Report(analysis.Diagnostic{
									Pos:     x.Pos(),
									Message: "direct function call `os.Exit()` in main package",
								})
								err = fmt.Errorf("direct function call `os.Exit()` in main package")
							}
						}
					}
					return true
				})
			}
		}
		return true
	})

	return err
}
