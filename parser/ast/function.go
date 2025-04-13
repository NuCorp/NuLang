package ast

import "github.com/LicorneSharing/GTL/optional"

type CallableFunc interface {
	FuncID() string
}

type FuncType struct {
	Arguments   []Type
	NamedArgs   map[int]string
	VariadicArg optional.Value[int]
	ReturnType  Type
}

type FuncDef struct {
	Name   string
	Header FuncType
	Body   any
}

type FuncExpr struct {
	Header FuncType
	Body   any
}

func (f FuncExpr) AsExpr() Expr { return f }
func (FuncExpr) FuncID() string { return "func:func" }

func (DotIdent) FuncID() string { return "func:named" }

type ArgBinding struct {
	Ordered []Expr
	Named   map[string]Expr
	// destructured ?
}

type ArgBindingElem interface {
	GetExpr() Expr
	IsNamed() bool
	IsDestructured() bool
}

type OrderArgBinding struct {
	Expr
}

func (o OrderArgBinding) GetExpr() Expr {
	return o
}

func (o OrderArgBinding) IsNamed() bool {
	return false
}

func (o OrderArgBinding) IsDestructured() bool {
	return false
}

type NamedArgBinding struct {
	Name DotIdent
	Expr Expr // may be nil
}

func (n NamedArgBinding) GetExpr() Expr {
	if n.Expr == nil {
		return n.Name
	}

	return n.Expr
}

func (n NamedArgBinding) IsNamed() bool {
	return true
}

func (n NamedArgBinding) IsDestructured() bool {
	return false
}

type DestructuredArgBinding struct {
	ArgBindingElem
}

func (d DestructuredArgBinding) IsDestructured() bool {
	return true
}

type FuncCall struct {
	From CallableFunc
	Args []ArgBindingElem
}

func (f FuncCall) AsExpr() Expr { return f }
