package parser

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
)

func Test_argParser_Parse(t *testing.T) {
	type testcase struct {
		name          string
		orderedParser ParserOf[*ast.OrderedArgElem]
		namedParser   ParserOf[*ast.NamedArgElem]
		scanner       fakeScanner
		wantErrs      Errors
		wantArg       ast.ArgElem
	}

	for _, tt := range []testcase{
		{
			name: "simple ordered",
			orderedParser: parserFuncFor[*ast.OrderedArgElem](func(s scan.Scanner, errors *Errors) *ast.OrderedArgElem {
				s.ConsumeTokenInfo()
				return &ast.OrderedArgElem{
					Expr: ast.IntExpr(42),
				}
			}),
			scanner: fakeScanner{
				tokens: []fakeScannerElem{
					literal(intToken(42)),
				},
			},
			wantArg: ast.ArgElem{
				Ordered: &ast.OrderedArgElem{
					Expr: ast.IntExpr(42),
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				errs = Errors{}
				got  = argParser{
					ordered: tt.orderedParser,
					named:   tt.namedParser,
				}.Parse(&tt.scanner, &tt.wantErrs)
			)

			tassert.Equal(t, tt.wantArg, got)
			tassert.Equal(t, tt.wantErrs, errs)
		})
	}
}

func Test_functionCallParser_ContinueParsing(t *testing.T) {
	from := ast.DotIdent{"callable"}

	type testcase struct {
		name      string
		argParser ParserOf[ast.ArgElem]
		scanner   fakeScanner
		want      ast.FuncCall
		wantErrs  Errors
	}
	for _, tt := range []testcase{
		{
			name: "simple",
			argParser: &fakeParserOf[ast.ArgElem]{
				Results: []ast.ArgElem{
					{
						Ordered: &ast.OrderedArgElem{ // 42 -> 1 token
							Expr: ast.IntExpr(42),
						},
					},
					{
						Named: &ast.NamedArgElem{ // *a: 42 -> 4 tokens
							Name: ast.DotIdent{"a"},
							Expr: ast.IntExpr(42),
						},
					},
				},
				Skip: []int{
					1,
					4,
				},
			},
			scanner: fakeScanner{
				tokens: []fakeScannerElem{
					token("("),
					/**/ literal(intToken(42)), token(","),
					/**/ token("*"), ident("a"), token(":"), literal(intToken(42)),
					token(")"),
				},
			},
			want: ast.FuncCall{
				From: from,
				Args: []ast.ArgElem{
					{
						Ordered: &ast.OrderedArgElem{
							Expr: ast.IntExpr(42),
						},
					},
					{
						Named: &ast.NamedArgElem{
							Name: ast.DotIdent{"a"},
							Expr: ast.IntExpr(42),
						},
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var errors Errors

			got := functionCallParser{
				args: listOf[parenthesesSurrounding, ast.ArgElem]{
					parser: tt.argParser,
				},
			}.ContinueParsing(from, &tt.scanner, &errors)

			tassert.Equal(t, tt.want, got)
			tassert.Equal(t, tt.wantErrs, errors)
		})
	}
}

func Test_namedArgParser_Parse(t *testing.T) {
	type testcase struct {
		name string
	}
	for _, tt := range []testcase{} {
		t.Run(tt.name, func(t *testing.T) {
			got := namedArgParser{
				// TODO
			}.Parse(nil, nil)

			tassert.Equal(t, nil, got) // TODO
		})
	}
}

func Test_orderedArgParser_Parse(t *testing.T) {
	type testcase struct {
		name string
	}
	for _, tt := range []testcase{} {
		t.Run(tt.name, func(t *testing.T) {
			got := orderedArgParser{
				// TODO
			}.Parse(nil, nil)

			tassert.Equal(t, nil, got) // TODO
		})
	}
}
