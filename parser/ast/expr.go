package ast

type Expr interface {
	ExprID() string
}

func (d DotIdent) ExprID() string {
	return "expr:ident"
}
