package ast

type LiteralExpr interface {
	Expr
	LiteralValue() any
}

type IntExpr int

func (i IntExpr) expr()             {}
func (i IntExpr) LiteralValue() any { return int(i) }
