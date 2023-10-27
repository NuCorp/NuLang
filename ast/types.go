package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
)

type Type interface {
	typeInterface()
}

type IdentType struct{ Ident }   // TODO use IdentType
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
