package ast

type LiteralExpr interface {
	Expr
	LiteralValue() any
}

type IntExpr int

func (i IntExpr) AsExpr() Expr      { return i }
func (i IntExpr) LiteralValue() any { return int(i) }
