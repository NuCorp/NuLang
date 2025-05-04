package ast

type LiteralExpr interface {
	Expr
	LiteralValue() any
}

type IntExpr int

func (i IntExpr) AsExpr() Expr      { return i }
func (i IntExpr) LiteralValue() any { return int(i) }

type StringExpr string

func (s StringExpr) AsExpr() Expr      { return s }
func (s StringExpr) LiteralValue() any { return string(s) }
