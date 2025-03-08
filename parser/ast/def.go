package ast

import "github.com/LicorneSharing/GTL/optional"

type Def interface {
	AsDef() Def
}

type Var struct {
	Name  string
	Type  Type
	Value Expr
}

func (v Var) AsDef() Def { return v }

type Const struct {
	IsConstexpr bool
	Name        string
	Type        Type
	Value       Expr
}

func (c Const) AsDef() Def { return c }

type TypeDef struct {
	Name      string
	Type      Type
	Extension Extension
	// With []TypeWith
}

func (t TypeDef) AsDef() Def { return t }

type Extension struct{}

type ExtensionDef struct {
	From      Type // Warning, must be a resolved type (not typeof(.))
	Extension Extension
}

func (e ExtensionDef) AsDef() Def { return e }

type CastKind int

const (
	Explicit = CastKind(iota)
	Implicit
	Delete
)

// type int as string = explicit: from {
type CastDef struct {
	From, To Type // Warning, can't be typeof(.)
	Kind     CastKind
	Body     optional.Value[any]
}

func (c CastDef) AsDef() Def { return c }

func (f FuncDef) AsDef() Def { return f }
