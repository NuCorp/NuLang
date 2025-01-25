package ast

type Expr interface {
	ExprID() string
}

func (d DotIdent) ExprID() string {
	return "expr:ident"
}

type IntExpr int

func (i IntExpr) ExprID() string {
	return "expr:int"
}

type AskOrOperator struct {
	Left  Expr
	Right Expr
}

func (o AskOrOperator) ExprID() string {
	return "expr:.??."
}

type AskOperator struct {
	Left Expr
}

func (o AskOperator) ExprID() string {
	return "expr:.?"
}

type ForceOperator struct {
	Left Expr
}

func (o ForceOperator) ExprID() string {
	return "expr:.!"
}
