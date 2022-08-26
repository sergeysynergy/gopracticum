// Package main Пакет реализует функционал статического анализатора кода на базе нескольких анализаторов:
// - стандартного набора анализаторов пакета `golang.org/x/tools/go/analysis/passes`;
// - всех анализаторов класса SA пакета staticcheck.io;
// - анализаторы QF1003, QF1010, S1000, S1001, ST1000 и ST1005 из пакета staticcheck.io;
// - анализатор из публичного пакета `github.com/kisielk/errcheck/errcheck`;
// - анализатор из публичного пакета `github.com/fatih/errwrap/errwrap`;
// - собственный анализатора `ExitAnalyzer`, который запрещает использовать прямой вызов os.Exit в функции main пакета main.
//
// Установка осуществляется командной `go install`.
//
// Выполнение команды `staticlint ./...` проанализирует рекурсивно все файлы и каталоги из вызванного места.
// Для анализа отдельного файла выполните `staticlint [имя_файла]`.
//
// По умолчанию применяются все анализаторы.
// Чтобы выбрать конкретные анализаторы, используйте флаг -NAME для каждого из них
// или -NAME=false для запуска всех анализаторов, которые явно не отключены.

package main

import (
	"fmt"
	"github.com/fatih/errwrap/errwrap"
	"github.com/kisielk/errcheck/errcheck"
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
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

func main() {
	// Использование пакета analysis требует минимальных усилий для создания multichecker с интерфейсом командной строки.
	// Создадим свой `multichecker` на базе пакет analysis.
	checksLen := len(staticcheck.Analyzers) + 50
	var myChecks = make([]*analysis.Analyzer, 0, checksLen)

	// Добавим в `multichecker` анализаторы из пакета analysis/passes.
	myChecks = append(myChecks, PassesChecks()...)

	// Добавим анализаторы из пакета analysis/passes.
	myChecks = append(myChecks, StaticChecks()...) // TODO: переделать через конфиг, вопрос как это сделать?
	myChecks = append(myChecks, SimpleChecks()...)
	myChecks = append(myChecks, StyleChecks()...)
	myChecks = append(myChecks, QuickFixChecks()...)

	// Добавим анализатор из открытого пакета `github.com/kisielk/errcheck/errcheck`
	// для проверки непроверенных ошибок в исходном коде go.
	myChecks = append(myChecks, errcheck.Analyzer)

	// Добавим анализатор из открытого пакета `https://github.com/fatih/errwrap`
	// для проверки оборачивания возвращаемых ошибок через директиву `%w`.
	myChecks = append(myChecks, errwrap.Analyzer)

	// Добавим собственный анализатор: запрещает использовать прямой вызов os.Exit в функции main пакета main.
	myChecks = append(myChecks, ExitAnalyzer)

	multichecker.Main(
		myChecks...,
	)
	//os.Exit(1)
}

// PassesChecks Подключает полный набор стандартных статических анализаторов пакета `golang.org/x/tools/go/analysis/passes`.
func PassesChecks() []*analysis.Analyzer {
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

// StaticChecks Подключает набор анализаторов класса SA из пакета `staticcheck`.
func StaticChecks() []*analysis.Analyzer {
	checks := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers))

	// Добавим все анализаторы из пакета `staticcheck` с префиксом SA
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			checks = append(checks, v.Analyzer)
		}
	}

	return checks
}

// SimpleChecks Подключает анализаторы пакета `simple` для проверки на:
// - S1000 Использование канала для приёма-передачи вместо одиночного case select.
// - S1001 Использование вызова на копирование объекта вместо цикла.
func SimpleChecks() []*analysis.Analyzer {
	checks := make([]*analysis.Analyzer, 0, 2)

	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1000" || v.Analyzer.Name == "S1001" {
			checks = append(checks, v.Analyzer)
		}
	}

	return checks
}

// StyleChecks Подключает анализаторы пакета `stylecheck` для проверки на:
// - ST1000 Неверное или отсутствующие комментирование пакета.
// - ST1005 Неверный формат описания ошибки.
func StyleChecks() []*analysis.Analyzer {
	checks := make([]*analysis.Analyzer, 0, 2)

	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" || v.Analyzer.Name == "ST1005" {
			checks = append(checks, v.Analyzer)
		}
	}

	return checks
}

// QuickFixChecks Подключает анализаторы пакета `quickfix` для проверки:
// - QF1003 Конвертация цепочки условий if/else-if в конструкцию switch.
// - QF1010 Конвертация слайса байт в строку при выводе.
func QuickFixChecks() []*analysis.Analyzer {
	checks := make([]*analysis.Analyzer, 0, 2)

	for _, v := range quickfix.Analyzers {
		if v.Analyzer.Name == "QF1003" || v.Analyzer.Name == "QF1010" {
			checks = append(checks, v.Analyzer)
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
	var err error

	ast.Inspect(file, func(node ast.Node) bool {
		if fl, ok := node.(*ast.File); ok {
			if fl.Name.Name != "main" {
				return false
			}
			for _, v := range fl.Decls {
				if mainFnc, ok := v.(*ast.FuncDecl); ok {
					if mainFnc.Name.Name != "main" {
						continue
					}
					// Рекурсивно проходим по элементам узла функция main: ищем прямое объявление os.Exit
					ast.Inspect(mainFnc, func(node ast.Node) bool {
						switch x := node.(type) {
						case *ast.SelectorExpr:
							if fmt.Sprint(x.X) == "os" && x.Sel.Name == "Exit" {
								err = fmt.Errorf("direct function call `os.Exit()` in main package")
								return false
							}
						}
						return true
					})
				}
			}
		}
		return true
	})

	return err
}
