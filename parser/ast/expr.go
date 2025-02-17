package ast

type Expr interface {
	expr()
}

func (d DotIdent) expr() {}

type TupleExpr []Expr

func (t TupleExpr) expr() {}

type AsTypeExpr struct {
	Forced bool
	Asked  bool
	From   Expr
	AsType Type
}

func (AsTypeExpr) expr() {}

type IsTypeExpr struct {
	Constexpr bool
	From      Expr
	IsType    Type
}

func (IsTypeExpr) expr() {}

type AddressOf struct {
	RealAddress bool
	Expr        Expr
}

func (AddressOf) expr() {}
