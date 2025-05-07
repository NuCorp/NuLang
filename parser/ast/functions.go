package ast

import "github.com/LicorneSharing/GTL/optional"

type CallableFunc interface {
	FuncID() string
}

type Argument struct {
	Name         string
	Type         Type
	DefaultValue Expr // nil if IsVariadic is true
	IsNamed      bool
	IsVariadic   bool
}

type FuncType struct {
	Arguments  []Argument
	ReturnType Type
}

type FuncDef struct {
	Name   optional.Value[string]
	Header FuncType
	Body   Scope
}

type FuncExpr struct {
	Header FuncType
	Body   any
}

func (f FuncExpr) AsExpr() Expr { return f }
func (FuncExpr) FuncID() string { return "func:func" }

func (DotIdent) FuncID() string { return "func:named" }

type (
	OrderedArgElem      = OrderedContainedElem
	NamedArgElem        = NamedContainedElem
	DestructuredArgElem = DestructuredContainedElem[ContainedElem]

	ArgElem = containedElem[OrderedArgElem, NamedArgElem, DestructuredArgElem]
)

type FuncCall struct {
	From CallableFunc
	Args []ArgElem
}

func (f FuncCall) AsExpr() Expr { return f }
func (f FuncCall) AsStmt() Stmt { return f }
