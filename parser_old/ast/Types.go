package ast

import (
	"fmt"
	"strings"

	"github.com/LicorneSharing/GTL/optional"
	"github.com/LicorneSharing/GTL/slices"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
)

type Types interface {
	String() string
	//Typename() string
}

type TypeValue interface {
	ConstexprType() Types
}

type TypeExpr struct {
	TypeIdent Ident
	OBrak     scan.TokenPos
	Type      Types
	CBrak     scan.TokenPos
}

func (te TypeExpr) String() string {
	return fmt.Sprintf("Type[%v]", te.Type)
}

func (te TypeExpr) ConstexprType() Types {
	return te.Type
}

type TypeExprType struct {
	TypeExpr TypeValue
	TypeKw   scan.TokenPos
}

func (t TypeExprType) String() string {
	return fmt.Sprintf("%v.type", t.TypeExpr)
}

type TypeOfExpr struct {
	TypeIdent Ident
	Constexpr bool
	OBrac     scan.TokenPos
	OfIdent   Ident
	Colon     scan.TokenPos
	Expr      Expr
	CBrac     scan.TokenPos
}

func (to TypeOfExpr) String() string {
	constexprIndicator := ""
	if to.Constexpr {
		constexprIndicator = "+"
	}

	return fmt.Sprintf("Type%v{Of: %v}", constexprIndicator, to.OfIdent)
}

func (to TypeOfExpr) Type() Types {
	return nil // to.Expr.InferredType()
}

type DotType struct {
	Idents []Ident
}

func (d DotType) String() string {
	return strings.Join(slices.Map(d.Idents, Ident.String), ".")
}

type StructType struct {
	StructKw    optional.Value[scan.TokenPos]
	OBrac       scan.TokenPos
	InnerStruct InnerStruct
}

func (ls StructType) String() string {
	opening := "{"

	if ls.StructKw.HasValue() {
		opening = "struct"
	}

	return opening + "{\n" + strings.ReplaceAll(ls.InnerStruct.String(), "\n", "\n\t") + "\n}"
}

type InnerStruct struct {
	Order          []Ident
	Fields         map[Ident]Types
	Gets           map[Ident]struct{}
	DefaultValue   map[Ident]Expr
	ComaSeparators map[Ident]scan.TokenPos
}

func (is InnerStruct) String() string {
	var str string

	for _, ident := range is.Order {
		var currentField string

		if _, ok := is.Gets[ident]; ok {
			currentField = "get "
		}

		currentField += ident.String()

		if typ, ok := is.Fields[ident]; ok {
			currentField += " " + typ.String()
		}

		if val, ok := is.DefaultValue[ident]; ok {
			currentField += " = " + val.String()
		}

		if _, ok := is.ComaSeparators[ident]; ok {
			currentField += ", "
		} else {
			currentField += "\n"
		}

		str += currentField
	}

	return strings.TrimSuffix(str, "\n")
}

type RefType struct {
	RefToken scan.TokenPos
	Of       Types
}

func (r RefType) String() string {
	return "&" + r.Of.String()
}

type PtrType struct {
	StarToken scan.TokenPos
	Of        Types
}

func (p PtrType) String() string {
	return "*" + p.Of.String()
}

type TupleType struct {
	OParen scan.TokenPos
	Types  []Types
	CParen scan.TokenPos
}

type UnknownType struct{}

func (UnknownType) String() string {
	return "øunknownø"
}
