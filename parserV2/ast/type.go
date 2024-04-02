package ast

import "github.com/DarkMiMolle/GTL/optional"

type TypeDef struct {
	TypeKw Keyword
	Name   Ident
	Type   NewTypeContent
}

type NewTypeContent interface {
	Ast
	newTypeContent()
}

type WithDefault struct {
	WithKw       Keyword
	DefaultKw    Keyword
	AssignSymbol Position

	Value  Ast                     // may be nil => Delete have value
	Delete optional.Value[Keyword] // may be nil => Value have value
}

type NewTypeFromExisting struct {
	ExistingType    Ast // TODO: Type // The type definition value
	Extension       optional.Value[ExtensionDef]
	WithConstraints optional.Value[WithClause]
	WithDefault     optional.Value[WithDefault]
}

func (f NewTypeFromExisting) newTypeContent() {}

type ImplementInterface struct {
	Colon          Position
	InterfaceNames []Ident
}

type WithClause struct {
	WithKw                Keyword
	OBrack                Position
	ConstraintsExpression []Ast // TODO: Expr
	CBrack                Position
}
