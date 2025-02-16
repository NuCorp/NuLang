package parser

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/NuCorp/NuLang/container"
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
)

func Test_structTypeParser_Parse(t *testing.T) {
	intTypeParser := parserFuncFor[ast.Type](func(s scan.Scanner, errors *Errors) ast.Type {
		s.ConsumeTokenInfo()
		return ast.NamedType{"int"}
	})

	testcases := []struct {
		name           string
		code           string
		haveTypeParser ParserOf[ast.Type]
		isInTypedef    bool

		wantStructType ast.StructType
		wantErrors     Errors
	}{
		{
			name:           "simple struct without getter",
			code:           "struct{a int, b int}",
			haveTypeParser: intTypeParser,

			wantStructType: ast.StructType{
				Fields: map[string]ast.Type{
					"a": ast.NamedType{"int"},
					"b": ast.NamedType{"int"},
				},
				GetFields:    make(container.Set[string]),
				DefaultValue: make(map[string]ast.Expr),
			},
			wantErrors: Errors{},
		},
		{
			name:           "multiple line struct",
			code:           "struct{\na int\nb int\n}",
			haveTypeParser: intTypeParser,
			wantStructType: ast.StructType{
				Fields: map[string]ast.Type{
					"a": ast.NamedType{"int"},
					"b": ast.NamedType{"int"},
				},
				GetFields:    make(container.Set[string]),
				DefaultValue: make(map[string]ast.Expr),
			},
			wantErrors: Errors{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				scanner = scan.Code(tt.code)
				errors  = Errors{}

				got = structTypeParser{typedef: tt.isInTypedef, typeParser: tt.haveTypeParser}.Parse(scanner, &errors)
			)

			tassert.Equal(t, tt.wantStructType, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}
