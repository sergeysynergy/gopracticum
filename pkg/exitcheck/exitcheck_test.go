package exitcheck

import (
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"testing"
)

func TestOsexit(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name: "Ok",
			src: `package may
func main() {
	os.Exit(0)
}
`,
			wantErr: false,
		},
		{
			name: "Not main package",
			src: `package may
func main() {
	kek := os.Exit(1)
}
`,
			wantErr: false,
		},
		{
			name: "No main function in package main",
			src: `package main

import "fmt"

func may() {
	kek := os.Exit(1)
    fmt.Println(kek)
}
`,
			wantErr: false,
		},
		{
			name: "Indirect os.Exit() call",
			src: `package main

import "fmt"

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic occurred:", err)
		}
	}()
	//_ := os.Exit(1)
}
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.src, 0)
			if err != nil {
				panic(err)
			}
			pass := &analysis.Pass{
				Analyzer: Analyzer,
				Files:    []*ast.File{f},
			}
			_, err = pass.Analyzer.Run(pass)

			if tt.wantErr {
				assert.Errorf(t, err, "direct function call `os.Exit()` in main package")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
