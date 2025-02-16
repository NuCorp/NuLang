package parser

import (
	"testing"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
)

type fakeParserOf[T any] struct {
	Skip    []int
	Results []T
	Errors  []map[int]string

	Current int
}

func (f *fakeParserOf[T]) Parse(s scan.Scanner, errors *Errors) T {
	defer func() {
		f.Current++
	}()

	for i := range f.Skip[f.Current] {
		if msg, ok := f.Errors[f.Current][i]; ok {
			errors.Set(s.CurrentPos(), msg)
		}

		s.ConsumeTokenInfo()
	}

	return f.Results[f.Current]
}

func Test_tupleExpr_Parse(t1 *testing.T) {
	testcases := []struct {
		name           string
		code           string
		withExprParser ParserOf[ast.Expr]

		wantTuple ast.TupleExpr
	}{
		{
			name: "single tuple",
			code: "(42)",
			withExprParser: parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			}),

			wantTuple: ast.TupleExpr{
				ast.IntExpr(42),
			},
		},
		{
			name: "multiple expr",
			code: "(42, 42)",
			withExprParser: parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			}),
			wantTuple: ast.TupleExpr{
				ast.IntExpr(42),
				ast.IntExpr(42),
			},
		},
		{
			name: "tuple of tuple",
			code: "((42, 42), 42)",
			withExprParser: &fakeParserOf[ast.Expr]{
				Skip: []int{
					5,
					1,
				},
				Results: []ast.Expr{
					ast.TupleExpr{ast.IntExpr(42), ast.IntExpr(42)},
					ast.IntExpr(42),
				},
			},
			wantTuple: ast.TupleExpr{
				ast.IntExpr(42),
				ast.IntExpr(42),
				ast.IntExpr(42),
			},
		},
	}
	for _, tt := range testcases {
		t1.Run(tt.name, func(t1 *testing.T) {

		})
	}
}
