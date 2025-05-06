package ast

import "github.com/LicorneSharing/GTL/optional"

type Def interface {
	Stmt
	AsDef() Def
}

type Var struct {
	Name  string
	Type  Type
	Value Expr
}

func (v Var) AsDef() Def   { return v }
func (v Var) AsStmt() Stmt { return v }

type Const struct {
	IsConstexpr bool
	Name        string
	Type        Type
	Value       Expr
}

func (c Const) AsDef() Def   { return c }
func (c Const) AsStmt() Stmt { return c }

type TypeDef struct {
	Name      string
	Type      Type // nil if Const is true
	Const     bool
	Extension Extension
	// With []TypeWith
}

func (t TypeDef) AsDef() Def   { return t }
func (t TypeDef) AsStmt() Stmt { return t }

type Extension struct{}

type ExtensionDef struct {
	From      Type // Warning, must be a resolved type (not typeof(.))
	Extension Extension
}

func (e ExtensionDef) AsDef() Def   { return e }
func (e ExtensionDef) AsStmt() Stmt { return e }

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

func (c CastDef) AsDef() Def   { return c }
func (c CastDef) AsStmt() Stmt { return c }

func (f FuncDef) AsDef() Def   { return f }
func (f FuncDef) AsStmt() Stmt { return f }
