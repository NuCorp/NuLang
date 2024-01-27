package ast

import "github.com/DarkMiMolle/GTL/optional"

type FunctionDef struct {
	FuncKw      Keyword
	Name        Ident
	Parameters  []Parameter
	ReturnType  optional.Value[Ast] // TODO: TypeExpr
	HasImplem   bool
	Body        []Ast
	ClosingBody optional.Value[Position] // }
}

/*
parameters:
func(a, b int) => Parameter(a, int), (b, int)

func (a int = expr) => Parameter(a, int, expr)

func(a ...int) => VariadicParameter(a, int)

func(*a int) => NamedParameter(a, int)

Rules:
	After Assignment => everything must be assigned or named only.
	After Variadic => everything must be named only.
	Only one Variadic.
*/

type Parameter interface {
	Ast
	parameter()
}

type SimpleParameter struct {
	Name  Ident
	Type  optional.Value[Ast] // TODO: TypeExpr
	Value optional.Value[Ast] // TODO: Expr
}

func (SimpleParameter) parameter() {}

type NamedParameter struct {
	Star Position
	*SimpleParameter
}

type VariadicParameter struct {
	Name Ident
	Type Ast // TODO: TypeExpr
}

func (VariadicParameter) parameter() {}
