package ast

type Expr interface {
	AsExpr() Expr
}

func (d DotIdent) AsExpr() Expr { return d }

type TupleExpr []Expr

func (t TupleExpr) AsExpr() Expr { return t }

type AsTypeExpr struct {
	Forced bool
	Asked  bool
	From   Expr
	AsType Type
}

func (a AsTypeExpr) AsExpr() Expr { return a }

type IsTypeExpr struct {
	Constexpr bool
	From      Expr
	IsType    Type
}

func (i IsTypeExpr) AsExpr() Expr { return i }

type AddressOf struct {
	RealAddress bool
	Expr        Expr
}

func (a AddressOf) AsExpr() Expr { return a }
