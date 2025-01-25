package ast

type Def interface {
	DefID() string
}

type Var struct {
	Name  string
	Type  Type
	Value Expr
}

func (v Var) DefID() string {
	return "def:var"
}

type Const struct {
	IsConstexpr bool
	Name        string
	Type        Type
	Value       Expr
}

func (Const) DefID() string {
	return "def:const"
}

type TypeDef struct {
	Name      string
	Type      Type
	Extension Extension
	// With []TypeWith
}

func (t TypeDef) DefID() string {
	return "def:type"
}

type Extension struct{}

type ExtensionDef struct {
	From      Type // Warning, must be a resolved type (not typeof(.))
	Extension Extension
}

func (ExtensionDef) DefID() string { return "def:extension" }

type FuncDef struct {
	Name      string
	Prototype FuncType
	Body      any
}

func (FuncDef) DefID() string { return "def:func" }
