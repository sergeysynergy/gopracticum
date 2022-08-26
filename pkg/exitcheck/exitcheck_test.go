package exitcheck

import (
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
			name:    "Ok",
			src:     `package main`,
			wantErr: false,
		},
		{
			name:    "Not main package",
			src:     `package shmain`,
			wantErr: true,
		},
		{
			name:    "No main function in package main",
			src:     `package main`,
			wantErr: true,
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
			run(pass)

			//if tt.wantErr {
			//	assert.Errorf(t, err, "direct function call `os.Exit()` in main package")
			//} else {
			//	assert.NoError(t, err)
			//}
		})
	}
}
