package ast

type Expr interface {
	ExprID() string
}

func (d DotIdent) ExprID() string {
	return "expr:ident"
}

type TupleExpr []Expr

func (t TupleExpr) ExprID() string { return "expr:tuple" }
