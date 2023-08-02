package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
)

type DottedType struct {
	DottedExpr
}

type LStructType struct {
	Opening    scanner.TokenPos
	Attributes []VarDef
	Getter     []bool // Getter.Length == Attribute.Length
	Ending     scanner.TokenPos
}

func (s LStructType) From() scanner.TokenPos {
	return s.Opening
}
func (s LStructType) To() scanner.TokenPos {
	return s.Ending
}
func (s LStructType) String() string {
	str := "struct{"
	for idx, attribute := range s.Attributes {
		if s.Getter[idx] {
			str += " get"
		}
		str += fmt.Sprintf(" %v;", attribute)
	}
	return str + " }"
}
