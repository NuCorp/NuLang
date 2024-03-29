package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"strings"
)

type VarList struct {
	Keyword scan.TokenPos
	count   int

	Vars []VarElem
}

func (v *VarList) From() scan.TokenPos {
	return v.Keyword
}
func (v *VarList) To() scan.TokenPos {
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

func (n *NamedDef) From() scan.TokenPos {
	return n.Name.from.FromPos()
}
func (n *NamedDef) To() scan.TokenPos {
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
	Star        scan.TokenPos
	OpenBrace   tokens.Token
	Elements    []BindingElement
	hasLeft     bool
	CloseBrace  tokens.Token
	AssignToken tokens.Token
	Value       Ast
}

func (*NameBinding) varElem() {}

func (n *NameBinding) From() scan.TokenPos {
	return n.Star
}
func (n *NameBinding) To() scan.TokenPos {
	return n.Value.To()
}
func (n *NameBinding) String() string {
	str := "*{"
	for _, element := range n.Elements {
		str += element.String() + ", "
	}
	str = strings.TrimSuffix(str, ", ")
	return str + fmt.Sprintf("} %v %v", n.AssignToken.String(), n.Value.String())
}

func (n *NameBinding) AddBindingElement(element BindingElement) {
	if _, isBindingLeft := element.(*BindingLeft); isBindingLeft {
		n.hasLeft = true
	}
	n.Elements = append(n.Elements, element)
}
func (n *NameBinding) HasBindingLeft() bool {
	return n.hasLeft
}

type NameBindingElem struct {
	VariableName  Ident
	Colon         tokens.Token
	AttributeName Ident
	BindingExpr   Ast // from AttributeName
	// linked value: for def checking
}

func (NameBindingElem) bindingElement() {}

func (n NameBindingElem) From() scan.TokenPos {
	return n.VariableName.From()
}
func (n NameBindingElem) To() scan.TokenPos {
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

type SubBinding struct {
	Opening       scan.TokenInfo
	Elements      []BindingElement
	hasLeft       bool
	Closing       tokens.Token
	Colon         tokens.Token
	AttributeName Ident // required
}

func (s *SubBinding) From() scan.TokenPos {
	return s.Opening.FromPos()
}
func (s *SubBinding) To() scan.TokenPos {
	return s.AttributeName.To()
}
func (s *SubBinding) String() string {
	str := s.Opening.Token().String()
	for _, element := range s.Elements {
		str += fmt.Sprintf("%v, ", element)
	}
	str = strings.TrimSuffix(str, ", ")
	return str + fmt.Sprintf("%v: .%v", s.Closing, s.AttributeName)
}

func (s *SubBinding) AddBindingElement(element BindingElement) {
	if _, isBindingLeft := element.(*BindingLeft); isBindingLeft {
		s.hasLeft = true
	}
	s.Elements = append(s.Elements, element)
}
func (s *SubBinding) HasBindingLeft() bool {
	return s.hasLeft
}

func (*SubBinding) bindingElement() {}

type BindingLeft struct {
	VariableName Ident
	Colon        tokens.Token
	Ellipsis     scan.TokenInfo
}

func (*BindingLeft) bindingElement() {}

func (b *BindingLeft) From() scan.TokenPos {
	return b.VariableName.From()
}
func (b *BindingLeft) To() scan.TokenPos {
	return b.Ellipsis.ToPos()
}
func (b *BindingLeft) String() string {
	return fmt.Sprintf("%v: ...", b.VariableName)
}

type BindingElement interface {
	Ast
	bindingElement()
}
