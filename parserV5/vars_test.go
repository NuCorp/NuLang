package parserV5

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV5/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
)

func Test_groupedVar_Parse(t *testing.T) {
	testcases := []struct {
		name string
		code string

		typeParser parserFuncFor[ast.Type]
		exprParser parserFuncFor[ast.Expr]

		expectVars []ast.Var
	}{
		{
			name: "single var typed",
			code: "a int",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			},
			expectVars: []ast.Var{
				{
					Name: "a",
					Type: ast.NamedType{"int"},
				},
			},
		},
		{
			name: "single var assigned",
			code: "a = 42",
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			expectVars: []ast.Var{
				{
					Name:  "a",
					Value: ast.IntExpr(42),
				},
			},
		},
		{
			name: "single var assigned and typed",
			code: "a float = 42",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"float"}
			},
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			expectVars: []ast.Var{
				{
					Name:  "a",
					Type:  ast.NamedType{"float"},
					Value: ast.IntExpr(42),
				},
			},
		},
		{
			name: "multiple var typed",
			code: "a, b int",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			},
			expectVars: []ast.Var{
				{
					Name: "a",
					Type: ast.NamedType{"int"},
				},
				{
					Name: "b",
					Type: ast.NamedType{"int"},
				},
			},
		},
		{
			name: "var assigned and vars typed",
			code: "a = 42, b, c int, d int = 42",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			},
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			expectVars: []ast.Var{
				{
					Name:  "a",
					Value: ast.IntExpr(42),
				},
				{
					Name: "b",
					Type: ast.NamedType{"int"},
				},
				{
					Name: "c",
					Type: ast.NamedType{"int"},
				},
				{
					Name:  "d",
					Type:  ast.NamedType{"int"},
					Value: ast.IntExpr(42),
				},
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				scanner = scan.Code(tt.code)
				errors  = Errors{}
			)

			vars := groupedVar{
				typeParser: tt.typeParser,
				expr:       tt.exprParser,
			}.Parse(scanner, &errors)

			tassert.Equal(t, tt.expectVars, vars)
		})
	}
}

func Test_nameBindingAssigned_Parse(t *testing.T) {
	testcases := []struct {
		name string
		code string

		subbindingOrderParser ParserOf[ast.OrderBindingAssign]
		exprParser            parserFuncFor[ast.Expr]

		wantNameBinding ast.NameBindingAssign
		wantErrors      Errors
	}{
		{
			name: "no sub-binding",
			code: "{a, b}",
			wantNameBinding: ast.NameBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.DotIdent{"b"},
				},
			},
		},
		{
			name: "name sub-binding",
			code: "{a, *{b}: .c}",
			wantNameBinding: ast.NameBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.NameBindingAssign{
						Elems: []ast.SubBinding{
							ast.DotIdent{"b"},
						},
					},
				},
				ToName: map[int]*ast.DotIdent{
					1: {"", "c"},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				parser = &nameBindingAssigned{
					expr: tt.exprParser,
				}
				scanner = scan.Code(tt.code)
				errors  = Errors{}
			)

			parser.subbinding = subbindingParser{
				namebindingAssign:  parser,
				orderbindingAssign: tt.subbindingOrderParser,
			}

			got := parser.Parse(scanner, &errors)

			tassert.Equal(t, tt.wantNameBinding, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}
