package ast

import "github.com/NuCorp/NuLang/utils/set"

type Expr interface {
	AsExpr() Expr
}

func (d DotIdent) AsExpr() Expr { return d }

type Elem = OrderedContainedElem
type DestructuredOrderedElem = DestructuredContainedElem[Elem]

// ListElem allows EXPR and EXPR...
type ListElem = containedElem[Elem, noneContainedElem, DestructuredOrderedElem]

func NewListElem(expr Expr) ListElem {
	return ListElem{
		Ordered: &OrderedContainedElem{
			Expr: expr,
		},
	}
}

func DestructuredListElem(l ListElem) ListElem {
	if l.Ordered == nil || l.Destructured != nil {
		return l
	}

	l.Destructured = &DestructuredOrderedElem{
		ContainedElem: *l.Ordered,
	}

	l.Ordered = nil

	return l
}

// TupleExpr can be : (EXPR, EXPR...)
type TupleExpr []ListElem

func (t TupleExpr) AsExpr() Expr { return t }

// ArrayExpr can be : [EXPR, EXPR...]
type ArrayExpr []ListElem

func (a ArrayExpr) AsExpr() Expr { return a }

type NamedElem = NamedContainedElem
type DestructuredNamedElem = DestructuredContainedElem[NamedElem]

type DictExpr struct {
	Keys         []Expr
	Values       map[int]Expr
	Destructured set.Set[int]
}

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
