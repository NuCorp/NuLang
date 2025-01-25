package ast

type LiteralExpr interface {
	Expr
	LiteralValue() any
}

type IntExpr int

func (i IntExpr) ExprID() string {
	return "expr:int"
}
func (i IntExpr) LiteralValue() any { return int(i) }
