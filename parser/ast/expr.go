package ast

import "github.com/LicorneSharing/GTL/optional"

type Expr interface {
	ExprID() string
}

func (d DotIdent) ExprID() string {
	return "expr:ident"
}

type TupleExpr []Expr

func (t TupleExpr) ExprID() string { return "expr:tuple" }

type AsTypeExpr struct {
	Forced bool
	Asked  bool
	From   Expr
	AsType Type
}

func (AsTypeExpr) ExprID() string { return "expr:as" }

type IsTypeExpr struct {
	Constexpr bool
	From      Expr
	IsType    Type
}

func (IsTypeExpr) ExprID() string { return "expr:is" }

type AddressOf struct {
	RealAddress bool
	Expr        Expr
}

func (AddressOf) ExprID() string { return "expr:&.:&(.)" }

type InitExpr struct {
	Type Type
	Name optional.Value[string]
	Args ArgBinding
}

func (InitExpr) ExprID() string { return "expr:init" }
