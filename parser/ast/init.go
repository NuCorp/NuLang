package ast

import "github.com/LicorneSharing/GTL/optional"

type InterfaceInitExpr struct {
	Name optional.Value[NamedType] // name of the type. No name => interface anonymous

	/*
		Methods is the field used to represent
			Error:{
				const Msg() => "msg"
			}
		or
			Error:{
				const Msg() string {
					return "msg"
				}
			}

		if there is only 1 FuncDef without name that means we are in the following pattern
			Error => "msg"
		or
			I(a int) => a
	*/
	Methods []FuncDef
}

func (i InterfaceInitExpr) AsExpr() Expr { return i }
func (InterfaceInitExpr) init()          {}

type ThrowIndicator int

const (
	NotThrowing ThrowIndicator = iota
	MayThrow
	MustThrow
)

type SimpleInitArg struct {
	Name         DotIdent
	Value        Expr
	Bool         optional.Value[bool]
	Destructured bool
}

type SimpleInitExpr struct {
	Type     Type
	MayThrow ThrowIndicator
	FromAs   optional.Value[Expr]
	Args     map[string]SimpleInitArg
}

func (s SimpleInitExpr) AsExpr() Expr { return s }
func (SimpleInitExpr) init()          {}

type NamedInitExpr struct {
	Type      Type
	Name      string
	MayThrow  bool
	Args      []Expr
	NamedArgs map[int]string
	BoolArgs  map[int]bool
	// Unstructured map[int]struct{}
	// NamedUnstructured map[int]struct{}
}

type InitExpr interface {
	Expr
	init()
}
