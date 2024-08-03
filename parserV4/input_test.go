package parserV4

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV4/ast"
	"testing"

	tassert "github.com/stretchr/testify/assert"
)

func Test_organizeBinaryOperator(t *testing.T) {
	exprs := []ast.Expr{
		ast.Literal[int]{Value: 0},
		ast.Literal[int]{Value: 1},
		ast.Literal[int]{Value: 2},
		ast.Literal[int]{Value: 3},
	}
	tests := []struct {
		name         string
		initialState ast.BinaryExpr
		want         ast.BinaryExpr
	}{
		{
			name: "success",
			initialState: ast.BinaryExpr{
				Left: exprs[0],
				Op:   "??",
				Right: &ast.BinaryExpr{
					Left: exprs[1],
					Op:   "+",
					Right: &ast.BinaryExpr{
						Left:  exprs[2],
						Op:    "-",
						Right: exprs[3],
					},
				},
			},
			want: ast.BinaryExpr{
				Left: &ast.BinaryExpr{
					Left:  exprs[0],
					Op:    "??",
					Right: exprs[1],
				},
				Op: "+",
				Right: &ast.BinaryExpr{
					Left:  exprs[2],
					Op:    "-",
					Right: exprs[3],
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			organizeBinaryOperator(&tt.initialState)
			tassert.Equal(t, tt.want, tt.initialState)
		})
	}
}
