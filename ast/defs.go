package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
)

type VarList struct {
	Keyword scanner.TokenPos
	count   int

	Vars []VarDef
}

func (v *VarList) From() scanner.TokenPos {
	return v.Keyword
}
func (v *VarList) To() scanner.TokenPos {
	return v.Vars[v.count].To()
}
func (v *VarList) AddVars(vars ...VarDef) {
	v.count += len(vars)
	v.Vars = append(v.Vars, vars...)
}
func (v *VarList) String() string {
	str := "var "
	for _, varElem := range v.Vars {
		str += varElem.String() + ", "
	}
	return strings.TrimSuffix(str, ", ")
}

type VarDef struct {
	Name   Value[string]
	Typ    Ast // TODO Type ?
	Assign tokens.Token
	Value  Ast // TODO Expr

}

func (v VarDef) From() scanner.TokenPos {
	return v.Name.from.FromPos()
}
func (v VarDef) To() scanner.TokenPos {
	if v.Value != nil {
		return v.To()
	}
	return v.Typ.To()
}
func (v VarDef) String() string {
	str := v.Name.Value
	if v.Typ != nil {
		str += fmt.Sprintf(" %v", v.Typ)
	}
	if v.Value != nil {
		str += fmt.Sprintf(" = %v", v.Value)
	}
	return str
}
