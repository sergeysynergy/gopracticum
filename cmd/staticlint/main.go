// Package main Пакет реализует функционал статического анализатора кода на базе нескольких анализаторов:
// - стандартного набора анализаторов пакета `golang.org/x/tools/go/analysis/passes`;
// - всех анализаторов класса SA пакета staticcheck.io;
// - анализатор [] из пакета staticcheck.io;
// - анализатор из публичного пакета ...;
// - анализатор из публичного пакета ...;
// - собственный анализатора `ExitAnalyzer`, который запрещает использовать прямой вызов os.Exit в функции main пакета main.

package main

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
	"strings"
)

func main() {
	// Использование пакета analysis требует минимальных усилий для создания multichecker с интерфейсом командной строки.
	// Создадим свой `multichecker` на базе пакет analysis.
	myChecks := []*analysis.Analyzer{
		ExitAnalyzer,
	}

	// Добавим в `multichecker` анализаторы из пакета analysis/passes.
	myChecks = append(myChecks, GetPasses()...)

	// TODO: переделать через конфиг
	//myChecks = append(myChecks, getStaticChecks()...)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	//appfile, err := os.Executable()
	//if err != nil {
	//	panic(err)
	//}
	//data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	//if err != nil {
	//	data, err = os.ReadFile(filepath.Join("./", Config))
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//var cfg ConfigData
	//if err = json.Unmarshal(data, &cfg); err != nil {
	//	panic(err)
	//}
	//
	//// определяем map подключаемых правил на основе файла конфигурации
	//checks := make(map[string]bool)
	//for _, v := range cfg.Staticcheck {
	//	checks[v] = true
	//}

	multichecker.Main(
		myChecks...,
	)
}

// GetPasses Подключает полный набор стандартных статических анализаторов пакета `golang.org/x/tools/go/analysis/passes`.
func GetPasses() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		findcall.Analyzer,
		httpresponse.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}
}

// Config — имя файла конфигурации.
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string
}

func getStaticChecks() []*analysis.Analyzer {
	checks := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers))

	// Добавим все анализаторы из пакета `staticcheck` с префиксом SA
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Name, "SA") {
			checks = append(checks, v)
		}
	}

	return checks
}

// ExitAnalyzer Запрещает использовать прямой вызов os.Exit в функции main пакета main.
var ExitAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for direct os.Exit call in main function of main package",
	Run:  run, // функция, которая отвечает за анализ исходного кода
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		err := ExitCheck(file)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func ExitCheck(file *ast.File) error {
	// Флаги не самый элегантный вариант решения задачи.
	// Но с учётом рекурсивного прохода по древовидному графу - рабочий.
	// Задаём флаг, является ли пакет main'ом.
	mainPackage := false
	// Задаём флаг, является ли функция main'ом в пакете main.
	mainFunc := false

	var err error

	ast.Inspect(file, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.File:
			if x.Name.Name == "main" {
				mainPackage = true
			} else {
				mainPackage = false
			}
		case *ast.FuncDecl:
			if x.Name.Name == "main" && mainPackage {
				mainFunc = true
			} else {
				mainFunc = false
			}
		case *ast.SelectorExpr:
			if mainPackage && mainFunc {
				if fmt.Sprint(x.X) == "os" && x.Sel.Name == "Exit" {
					err = fmt.Errorf("direct function call `os.Exit()` in main package")
					return false
				}
			}
		}

		// Другим вариантом так и не понял, как обойти узлы...
		//if fl, ok := n.(*ast.File); ok {
		//	if fl.Name.Name != "main" {
		//		return true
		//	}
		//	for _, v := range fl.Decls {
		//		if mainFnc, ok := v.(*ast.FuncDecl); ok {
		//			if mainFnc.Name.Name != "main" {
		//				continue
		//			}
		//			for _, j := range mainFnc.Body.List {
		//				fmt.Println(j)
		//			}
		//		}
		//	}
		//}

		return true
	})

	return err
}
