package ast

import "github.com/LicorneSharing/GTL/optional"

type Expr interface {
	expr()
}

func (d DotIdent) expr() {}

type TupleExpr []Expr

func (t TupleExpr) expr() {}

type AsTypeExpr struct {
	Forced bool
	Asked  bool
	From   Expr
	AsType Type
}

func (AsTypeExpr) expr() {}

type IsTypeExpr struct {
	Constexpr bool
	From      Expr
	IsType    Type
}

func (IsTypeExpr) expr() {}

type AddressOf struct {
	RealAddress bool
	Expr        Expr
}

func (AddressOf) expr() {}

type InterfaceInitExpr struct {
	Name optional.Value[NamedType]

	/*
		Methods is the field used to represent
			Error {
				const Msg() => "msg"
			}
		or
			Error {
				const Msg() string {
					return "msg"
				}
			}
	*/
	Methods []FuncDef

	/*
		DirectMethod is the field used to represent
			Error => "msg"
		or
			Error {
				return "msg"
			}
	*/
	DirectMethod any
}

func (InterfaceInitExpr) expr() {}
func (InterfaceInitExpr) init() {}

type ClassicInitExpr struct {
	Type     Type
	MayThrow bool
	Named    optional.Value[string]
	FromAs   optional.Value[Expr]
	Args     map[string]Expr
	BoolArgs map[string]bool
}

func (ClassicInitExpr) expr() {}
func (ClassicInitExpr) init() {}

type InitExpr interface {
	Expr
	init()
}
