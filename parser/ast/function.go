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

type FuncCall struct {
	From CallableFunc
	Args ArgBinding
}

func (f FuncCall) AsExpr() Expr { return f }
