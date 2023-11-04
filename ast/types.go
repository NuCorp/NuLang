package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type Type interface {
	Ast
	typeInterface()
}

type IdentType struct {
	Ident
}

func (IdentType) typeInterface() {}

type DottedType struct {
	DottedExpr
}

func (DottedType) typeInterface() {}

type AnonymousStructType struct {
	Opening    scan.TokenPos
	Attributes []*NamedDef
	Getter     []bool // Getter.Length == Attribute.Length
	Ending     scan.TokenPos
}

func (AnonymousStructType) typeInterface() {}
func (s AnonymousStructType) From() scan.TokenPos {
	return s.Opening
}
func (s AnonymousStructType) To() scan.TokenPos {
	return s.Ending
}
func (s AnonymousStructType) String() string {
	str := "struct{"
	for idx, attribute := range s.Attributes {
		if s.Getter[idx] {
			str += " get"
		}
		str += fmt.Sprintf(" %v;", attribute)
	}
	return str + " }"
}

type TypeOf struct {
	Typeof  scan.TokenInfo
	OParent tokens.Token
	Static  tokens.Token
	Expr    Ast
	CParent scan.TokenInfo
}

func (TypeOf) typeInterface() {}

func (t TypeOf) From() scan.TokenPos {
	return t.Typeof.FromPos()
}
func (t TypeOf) To() scan.TokenPos {
	return t.CParent.ToPos()
}

func (t TypeOf) String() string {
	static := "+"
	if t.Static != tokens.PLUS {
		static = ""
	}
	return "typeof" + static + "(" + t.Expr.String() + ")"
}
