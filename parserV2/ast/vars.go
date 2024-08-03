package ast

import (
	"github.com/DarkMiMolle/GTL/optional"
)

// VarDeclaration is for variable declaration using the `var` keyword
type VarDeclaration struct {
	VarKeyword Keyword
	Variables  []Variable
}

type Variable struct {
	Name        Ident
	DeclareOnly bool
	Type        optional.Value[Ast]
	Value       optional.Value[Ast]
}

type declarationElem interface {
	Ast
	declarationElem()
}

type VarElem interface {
	declarationElem
}

type SingleVarDeclaration struct {
	Name        Ident
	DeclareOnly bool                // ? after the name
	Type        optional.Value[Ast] // TODO replace Ast by TypeExpr
	Value       optional.Value[Ast] // TODO replace Ast by Expr
}

func (SingleVarDeclaration) declarationElem() {}

// SingleVar methods

//

type assignmentBinding[T BindingElement] struct {
	StarSymbol optional.Value[Position]
	Opening    Position

	Element []T

	Optional optional.Value[Position] // element are optional so if binding is not possible => no crash
	Forced   optional.Value[Position] // unsafe binding so if binding is not possible => crash
}

type NameBinding struct {
	assignmentBinding[NameBindingElement]
	Value Ast // TODO replace Ast by Expr
}

type OrderBinding struct {
	assignmentBinding[OrderBindingElement]
	Value optional.Value[Ast] // TODO replace Ast by Expr
}

func (assignmentBinding[T]) declarationElem() {}

func (assignmentBinding[T]) bindingElement()      {}
func (assignmentBinding[T]) orderBindingElement() {}
func (assignmentBinding[T]) nameBindingElement()  {}

type BindingElement interface {
	bindingElement()
}

type NameBindingElement interface {
	BindingElement
	nameBindingElement()
}

type OrderBindingElement interface {
	BindingElement
	orderBindingElement()
}

type singleBindingElem struct {
	Elem Ident
	Cast optional.Value[CastBinding]
	Ref  optional.Value[RefBinding]
}

func (singleBindingElem) bindingElement() {}

type CastBinding struct {
	AsKeyword Keyword
	Type      Ast // TODO replace Ast by TypeExpr
}

type RefBinding struct {
	RefSymbol Position
}

type SingleOrderBindingElem struct {
	singleBindingElem
}

func (SingleOrderBindingElem) orderBindingElement() {}

type SingleNameBindingElem struct {
	singleBindingElem
	Rename optional.Value[Ident]
}

func (SingleNameBindingElem) nameBindingElement() {}

type LeftBinding struct {
	Elem       Ident
	LeftSymbol Position
}

func (LeftBinding) bindingElement()      {}
func (LeftBinding) orderBindingElement() {}
func (LeftBinding) nameBindingElement()  {}
