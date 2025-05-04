package parser

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
)

func Test_tupleExpr_Parse(t *testing.T) {
	testcases := []struct {
		name       string
		scanner    fakeScanner
		listParser ParserOf[ast.ListElem]

		wantTuple ast.TupleExpr
	}{
		{
			name: "single tuple",
			scanner: fakeScanner{
				tokens: []fakeScannerElem{
					token("("), literal(intToken(42)), token(")"),
				},
			},
			listParser: parserFuncFor[ast.ListElem](func(scanner scan.Scanner, errors *Errors) ast.ListElem {
				scanner.ConsumeTokenInfo()
				return ast.NewListElem(ast.IntExpr(42))
			}),

			wantTuple: ast.TupleExpr{ast.NewListElem(ast.IntExpr(42))},
		},
		{
			name: "multiple expr",
			scanner: fakeScanner{
				tokens: []fakeScannerElem{ // (42, 42)
					token("("), literal(intToken(42)), token(","), literal(intToken(42)), token(")"),
				},
			},
			listParser: parserFuncFor[ast.ListElem](func(scanner scan.Scanner, errors *Errors) ast.ListElem {
				scanner.ConsumeTokenInfo()
				return ast.ListElem{Ordered: &ast.OrderedContainedElem{ast.IntExpr(42)}}
			}),
			wantTuple: ast.TupleExpr{
				ast.NewListElem(ast.IntExpr(42)),
				ast.NewListElem(ast.IntExpr(42)),
			},
		},
		{
			name: "tuple of tuple",
			scanner: fakeScanner{ // (42, (42, 42))
				tokens: []fakeScannerElem{
					token("("), literal(intToken(42)), token(","), token("("), literal(intToken(42)), token(","), literal(intToken(42)), token(")"), token(")"),
				},
			},
			listParser: &fakeParserOf[ast.ListElem]{
				T: t,
				Skip: []int{
					1,
					5,
				},
				Results: []ast.ListElem{
					ast.NewListElem(ast.IntExpr(42)),
					ast.NewListElem(
						ast.TupleExpr{
							ast.NewListElem(ast.IntExpr(42)),
							ast.NewListElem(ast.IntExpr(42)),
						},
					),
				},
			},
			wantTuple: ast.TupleExpr{
				ast.NewListElem(ast.IntExpr(42)),
				ast.NewListElem(ast.IntExpr(42)),
				ast.NewListElem(ast.IntExpr(42)),
			},
		},
		{
			name: "tuple with destructured binding",
			scanner: fakeScanner{ // (42, a...)
				tokens: []fakeScannerElem{
					token("("), literal(intToken(42)), token(","), ident("a"), token("..."), token(")"),
				},
			},
			listParser: &fakeParserOf[ast.ListElem]{
				Results: []ast.ListElem{
					ast.NewListElem(ast.IntExpr(42)),
					ast.DestructuredListElem(ast.NewListElem(ast.DotIdent{"a"})),
				},
				Skip: []int{
					1,
					2,
				},
			},
			wantTuple: ast.TupleExpr{
				ast.NewListElem(ast.IntExpr(42)),
				ast.DestructuredListElem(ast.NewListElem(ast.DotIdent{"a"})),
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t1 *testing.T) {
			var (
				errs = Errors{}
				got  = tupleExpr{expr: tt.listParser}.Parse(&tt.scanner, &errs)
			)

			tassert.Equal(t1, tt.wantTuple, got)
		})
	}
}

func Test_arrayExpr_Parse(t *testing.T) {
	type testcase struct {
		name string

		scanner    fakeScanner
		exprParser ParserOf[ast.ListElem]

		wantArray  ast.ArrayExpr
		wantErrors Errors
	}

	for _, tt := range []testcase{
		{
			name: "empty array",
			scanner: fakeScanner{
				tokens: []fakeScannerElem{
					token("["), token("]"),
				},
			},
			exprParser: parserFuncFor[ast.ListElem](func(scan.Scanner, *Errors) ast.ListElem {
				t.Helper()
				t.Fatalf("should not reach here")
				return ast.ListElem{}
			}),
			wantArray: ast.ArrayExpr(nil),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				errors Errors
				got    = arrayExpr{expr: tt.exprParser}.Parse(&tt.scanner, &errors)
			)

			tassert.Equal(t, tt.wantArray, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}

func Test_dictExpr_Parse(t *testing.T) {
	type testcase struct {
		name        string
		scanner     fakeScanner
		exprParser  ParserOf[ast.Expr]
		identParser ParserOf[ast.DotIdent]

		wantDict ast.DictExpr
	}

	for _, tt := range []testcase{
		{
			name: "simple",
			scanner: fakeScanner{
				tokens: []fakeScannerElem{
					token("["),
					/**/ literal(stringToken("str")), token(":"), literal(intToken(42)),
					token("]"),
				},
			},
			exprParser: &fakeParserOf[ast.Expr]{
				T: t,
				Results: []ast.Expr{
					ast.StringExpr("str"),
					ast.IntExpr(42),
				},
				Skip: []int{1, 1},
			},
			wantDict: ast.DictExpr{
				Keys: []ast.Expr{
					ast.StringExpr("str"),
				},
				Values: map[int]ast.Expr{
					0: ast.IntExpr(42),
				},
			},
		},
		{
			name: "with ident key",
			scanner: fakeScanner{
				tokens: []fakeScannerElem{
					token("["),
					/* *a: 42, */ token("*"), ident("a"), token(":"), literal(intToken(42)), token(","),
					/* *b, */ token("*"), ident("b"), token(","),
					/* *c.d */ token("*"), ident("c"), token("."), ident("d"),
					token("]"),
				},
			},
			exprParser: parserFuncFor[ast.Expr](func(s scan.Scanner, e *Errors) ast.Expr {
				return ast.IntExpr(s.ConsumeTokenInfo().Value().(uint))
			}),
			identParser: &fakeParserOf[ast.DotIdent]{
				T: t,
				Results: []ast.DotIdent{
					{"a"},
					{"b"},
					{"c", "d"},
				},
				Skip: []int{1, 1, 3},
			},

			wantDict: ast.DictExpr{
				Keys: []ast.Expr{
					ast.StringExpr("a"),
					ast.StringExpr("b"),
					ast.StringExpr("d"),
				},
				Values: map[int]ast.Expr{
					0: ast.IntExpr(42),
					1: ast.DotIdent{"b"},
					2: ast.DotIdent{"c", "d"},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				errors = Errors{}
				got    = dictExpr{
					expr:  tt.exprParser,
					ident: tt.identParser,
				}.Parse(&tt.scanner, &errors)
			)
			tassert.Equal(t, tt.wantDict, got)
		})
	}
}
