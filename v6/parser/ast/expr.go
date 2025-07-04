package ast

type Expr interface {
	Node
	AsExpr() Expr
}
