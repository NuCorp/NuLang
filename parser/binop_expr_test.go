package parser

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func mustGetBinaryOperator(t tokens.Token) ast.BinaryOperator {
	op, ok := ast.GetBinopOperator(t)

	if !ok {
		panic("not a binary operator")
	}

	return op
}

func Test_binop_ContinueParsing(t *testing.T) {
	testcases := []struct {
		name     string
		code     string
		fromExpr ast.Expr

		withExprParser ParserOf[ast.Expr]

		wantBinop  ast.BinopExpr
		wantErrors Errors
	}{
		{
			name:     "single binop",
			code:     "+ 12",
			fromExpr: ast.IntExpr(42),
			withExprParser: parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(12)
			}),
			wantBinop: ast.BinopExpr{
				Left:  ast.IntExpr(42),
				Op:    mustGetBinaryOperator(tokens.PLUS),
				Right: ast.IntExpr(12),
			},
		},
		{
			name:     "single binop on multiple line",
			code:     "+ \n\t 12",
			fromExpr: ast.IntExpr(42),
			withExprParser: parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(12)
			}),
			wantBinop: ast.BinopExpr{
				Left:  ast.IntExpr(42),
				Op:    mustGetBinaryOperator(tokens.PLUS),
				Right: ast.IntExpr(12),
			},
		},
		{
			name:     "multiple binop",
			code:     "+ 12 * 2 - a ?? 28 / 3",
			fromExpr: ast.IntExpr(42),
			withExprParser: parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
				token := scanner.ConsumeTokenInfo()
				if token.Token().IsLiteral() {
					return ast.IntExpr(token.Value().(uint))
				}

				return ast.DotIdent{token.Value().(string)}
			}),
			wantBinop: ast.BinopExpr{ // -
				Left: ast.BinopExpr{ // +
					Left: ast.IntExpr(42),
					Op:   mustGetBinaryOperator(tokens.PLUS),
					Right: ast.BinopExpr{ // *
						Left:  ast.IntExpr(12),
						Op:    mustGetBinaryOperator(tokens.TIME),
						Right: ast.IntExpr(2),
					},
				},
				Op: mustGetBinaryOperator(tokens.MINUS),
				Right: ast.BinopExpr{ // /
					Left: ast.BinopExpr{ // ??
						Left:  ast.DotIdent{"a"},
						Op:    mustGetBinaryOperator(tokens.ASKOR),
						Right: ast.IntExpr(28),
					},
					Op:    mustGetBinaryOperator(tokens.DIV),
					Right: ast.IntExpr(3),
				},
			},
		},
		{
			name:     "multiple binop 2",
			code:     "* 12 + 2 / a ?? 28 - 3",
			fromExpr: ast.IntExpr(42),
			withExprParser: parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
				token := scanner.ConsumeTokenInfo()
				if token.Token().IsLiteral() {
					return ast.IntExpr(token.Value().(uint))
				}

				return ast.DotIdent{token.Value().(string)}
			}),
			wantBinop: ast.BinopExpr{ // -
				// 42 * 12 + 2 / a ?? 28
				Left: ast.BinopExpr{ // +
					Left: ast.BinopExpr{ // *
						Left:  ast.IntExpr(42),
						Op:    mustGetBinaryOperator(tokens.TIME),
						Right: ast.IntExpr(12),
					},
					Op: mustGetBinaryOperator(tokens.PLUS),
					// 2 / a ?? 28
					Right: ast.BinopExpr{ // /
						Left: ast.IntExpr(2),
						Op:   mustGetBinaryOperator(tokens.DIV),
						Right: ast.BinopExpr{
							Left:  ast.DotIdent{"a"},
							Op:    mustGetBinaryOperator(tokens.ASKOR),
							Right: ast.IntExpr(28),
						},
					},
				},
				Op:    mustGetBinaryOperator(tokens.MINUS),
				Right: ast.IntExpr(3),
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				parser  = binop{expr: tt.withExprParser}
				scanner = scan.Code(tt.code)
				errors  = Errors{}
				got     = parser.ContinueParsing(tt.fromExpr, scanner, &errors)
			)

			tassert.Equal(t, tt.wantBinop, got, "wanted: %v\ngot   : %v", tt.wantBinop, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}

func Test_fixeBinop(t *testing.T) {
	var (
		validExpr = ast.IntExpr(42)
		plus      = mustGetBinaryOperator(tokens.PLUS)
	)
	testcases := []struct {
		name      string
		fromBinop *ast.BinopExpr
		wantBinop ast.BinopExpr
	}{
		{
			name: "single binop",
			fromBinop: &ast.BinopExpr{
				Left:  ast.IntExpr(42),
				Op:    mustGetBinaryOperator(tokens.PLUS),
				Right: ast.IntExpr(42),
			},
			wantBinop: ast.BinopExpr{
				Left:  ast.IntExpr(42),
				Op:    mustGetBinaryOperator(tokens.PLUS),
				Right: ast.IntExpr(42),
			},
		},
		{
			name: "multiple binops",
			fromBinop: &ast.BinopExpr{
				Left: &ast.BinopExpr{
					Left: &ast.BinopExpr{
						Left: &ast.BinopExpr{
							Left:  validExpr,
							Op:    plus,
							Right: validExpr,
						},
						Op:    plus,
						Right: validExpr,
					},
					Op: plus,
					Right: &ast.BinopExpr{
						Left: validExpr,
						Op:   plus,
						Right: &ast.BinopExpr{
							Left:  validExpr,
							Op:    plus,
							Right: validExpr,
						},
					},
				},
				Op: plus,
				Right: &ast.BinopExpr{
					Left: &ast.BinopExpr{
						Left:  validExpr,
						Op:    plus,
						Right: validExpr,
					},
					Op: plus,
					Right: &ast.BinopExpr{
						Left: validExpr,
						Op:   plus,
						Right: &ast.BinopExpr{
							Left:  validExpr,
							Op:    plus,
							Right: validExpr,
						},
					},
				},
			},
			wantBinop: ast.BinopExpr{
				Left: ast.BinopExpr{
					Left: ast.BinopExpr{
						Left: ast.BinopExpr{
							Left:  validExpr,
							Op:    plus,
							Right: validExpr,
						},
						Op:    plus,
						Right: validExpr,
					},
					Op: plus,
					Right: ast.BinopExpr{
						Left: validExpr,
						Op:   plus,
						Right: ast.BinopExpr{
							Left:  validExpr,
							Op:    plus,
							Right: validExpr,
						},
					},
				},
				Op: plus,
				Right: ast.BinopExpr{
					Left: ast.BinopExpr{
						Left:  validExpr,
						Op:    plus,
						Right: validExpr,
					},
					Op: plus,
					Right: ast.BinopExpr{
						Left: validExpr,
						Op:   plus,
						Right: ast.BinopExpr{
							Left:  validExpr,
							Op:    plus,
							Right: validExpr,
						},
					},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			got := fixeBinop(tt.fromBinop)

			tassert.Equal(t, tt.wantBinop, got, "wanted: %v\ngot   :%v", tt.wantBinop, got)
		})
	}
}

func checkPriorities(b *ast.BinopExpr) bool {
	res := true

	if right, ok := b.Right.(ast.BinopExpr); ok {
		res = res &&
			binopPriorities[b.Op.BinaryOpToken()] > binopPriorities[right.Op.BinaryOpToken()]
		checkPriorities(&right)
	}

	if left, ok := b.Left.(ast.BinopExpr); ok {
		res = res &&
			binopPriorities[b.Op.BinaryOpToken()] > binopPriorities[left.Op.BinaryOpToken()]
		checkPriorities(&left)
	}

	return res
}

func Test_organizeBinaryOperator(t *testing.T) {
	// generate random binop and check that every time priorities are ok for every branch
}
