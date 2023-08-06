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

	Vars []VarElem
}

func (v *VarList) From() scanner.TokenPos {
	return v.Keyword
}
func (v *VarList) To() scanner.TokenPos {
	return v.Vars[v.count].To()
}
func (v *VarList) AddVars(vars ...VarElem) {
	v.count += len(vars)
	v.Vars = append(v.Vars, vars...)
}
func (v *VarList) String() string {
	str := "var "
	for _, varElem := range v.Vars {
		str += fmt.Sprintf("%v, ", varElem)
	}
	return strings.TrimSuffix(str, ", ")
}

type VarElem interface {
	Ast
	varElem()
}

type NamedDef struct {
	Name   Value[string]
	Typ    Ast // TODO Type ?
	Assign tokens.Token
	Value  Ast // TODO Expr
}

func (n *NamedDef) From() scanner.TokenPos {
	return n.Name.from.FromPos()
}
func (n *NamedDef) To() scanner.TokenPos {
	if n.Value != nil {
		return n.To()
	}
	return n.Typ.To()
}
func (n *NamedDef) String() string {
	str := n.Name.Value
	if n.Typ != nil {
		str += fmt.Sprintf(" %v", n.Typ)
	}
	if n.Value != nil {
		str += fmt.Sprintf(" = %v", n.Value)
	}
	return str
}

func (*NamedDef) varElem() {}

type NameBinding struct {
	Star        scanner.TokenPos
	OpenBrace   tokens.Token
	Elements    []*NameBindingElem
	Left        *BindingLeft
	CloseBrace  tokens.Token
	AssignToken tokens.Token
	Value       Ast
}

func (NameBinding) varElem() {}

func (n NameBinding) From() scanner.TokenPos {
	return n.Star
}
func (n NameBinding) To() scanner.TokenPos {
	return n.Value.To()
}
func (n NameBinding) String() string {
	str := "*{"
	for _, element := range n.Elements {
		str += element.String() + ", "
	}
	str = strings.TrimSuffix(str, ", ")
	if n.Left != nil {
		str += fmt.Sprintf(", %v", n.Left)
	}
	return str + fmt.Sprintf("} %v %v", n.AssignToken.String(), n.Value.String())
}

type NameBindingElem struct {
	VariableName  Ident
	Colon         tokens.Token
	AttributeName Ident
	BindingExpr   Ast // from AttributeName
	// linked value: for def checking
}

func (NameBindingElem) bindingElement() {}

func (n NameBindingElem) From() scanner.TokenPos {
	return n.VariableName.From()
}
func (n NameBindingElem) To() scanner.TokenPos {
	if n.Colon == tokens.NoInit {
		return n.VariableName.To()
	}
	return n.BindingExpr.To()
}
func (n NameBindingElem) String() string {
	str := n.VariableName.String()
	if n.Colon != tokens.NoInit {
		str += fmt.Sprintf(": %v", n.BindingExpr)
	}
	return str
}

type BindingLeft struct {
	VariableName Ident
	Colon        tokens.Token
	Ellipsis     scanner.TokenInfo
}

func (*BindingLeft) bindingElement() {}

func (b *BindingLeft) From() scanner.TokenPos {
	return b.VariableName.From()
}
func (b *BindingLeft) To() scanner.TokenPos {
	return b.Ellipsis.ToPos()
}
func (b *BindingLeft) String() string {
	return fmt.Sprintf("%v: ...", b.VariableName)
}

type BindingElement interface {
	Ast
	bindingElement()
}
