package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
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
	Opening    scanner.TokenPos
	Attributes []*NamedDef
	Getter     []bool // Getter.Length == Attribute.Length
	Ending     scanner.TokenPos
}

func (AnonymousStructType) typeInterface() {}
func (s AnonymousStructType) From() scanner.TokenPos {
	return s.Opening
}
func (s AnonymousStructType) To() scanner.TokenPos {
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
