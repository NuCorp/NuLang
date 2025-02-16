package ast

import (
	"fmt"

	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

//

type VarElem Ast

type VarList struct {
	Kw   scan.TokenPos
	Vars []VarElem
}

func (v VarList) String() string {
	str := "var "
	for _, var_ := range v.Vars {
		str += fmt.Sprint(var_)
	}
	return str
}
func (v VarList) CodePos() scan.TokenPos {
	return v.Kw
}

type DefaultedVar struct {
	Names []*Ident
	Type  any // TypeExpr
}

func (d DefaultedVar) String() string {
	str := ""
	for i, name := range d.Names {
		str += name.Name
		if i != len(d.Names)-1 {
			str += ", "
		}
	}
	return str + " " + fmt.Sprint(d.Type)
}
func (d DefaultedVar) CodePos() scan.TokenPos {
	return d.Names[0].CodePos()
}

type AssignedVar struct {
	Name     *Ident
	Type     Types
	Assigned scan.TokenPos
	Value    Expr
}

func (a AssignedVar) CodePos() scan.TokenPos {
	return a.Name.CodePos()
}
func (a AssignedVar) String() string {
	str := a.Name.String()
	if a.Type != nil {
		str += a.Type.String()
	}
	return str + " = ø" // convert to: str + " = " + a.Value.String()
}

type BindingElement interface {
	Ast
	bindingElement()
}

type BindingIdent struct {
	Ident   *Ident
	Ask     tokens.Token // either ? or ??
	ValueOr Expr         // optional value possible when Ask == tokens.AskOr
}

func (bi BindingIdent) String() string {
	str := bi.Ident.Name
	if bi.Ask.IsOneOf(tokens.ASK, tokens.ASKOR) {
		str += bi.Ask.String()
	}
	if bi.ValueOr != nil {

	}
	return str
}
func (bi BindingIdent) CodePos() scan.TokenPos {
	return bi.Ident.CodePos()
}

func (BindingIdent) bindingElement() {}

type InvalidBindingElement struct{} // info about the error here

func (i InvalidBindingElement) bindingElement() {}
func (i InvalidBindingElement) String() string {
	return "INVALID"
}
func (i InvalidBindingElement) CodePos() scan.TokenPos {
	return scan.TokenPos{}
}

type SubBinding interface {
	BindingElement
	subBinding()
}

type Destructure struct {
	Name     *Ident
	Ellipsis scan.TokenPos
}

func (d Destructure) String() string {
	return d.Name.String() + "..."
}
func (d Destructure) CodePos() scan.TokenPos {
	return d.Name.CodePos()
}

type subBinding[T any] struct {
	Open  scan.TokenPos
	Elems []T
	Left  *Destructure
	Close scan.TokenPos
}

func (subBinding[T]) subBinding() {}

func (subBinding[T]) bindingElement() {}

func (s subBinding[T]) String() string {
	var str string
	for i, elem := range s.Elems {
		str += fmt.Sprint(elem)
		if i != len(s.Elems)-1 {
			str += ", "
		}
	}
	switch any(new(T)).(type) {
	case *NameBindingAssociation:
		return "{ " + str + " }"
	default:
		return "[" + str + "]"
	}
}
func (s subBinding[T]) CodePos() scan.TokenPos {
	return s.Open
}

type NameBindingValueRef interface {
	Ast
	nameBindingValueRef()
}

func (*Ident) nameBindingValueRef() {}

type Literal[T any] struct {
	Pos   scan.TokenPos
	Value T
}

func (l Literal[T]) IsConstexpr() bool {
	return true
}

func (Literal[T]) nameBindingValueRef() {}
func (l Literal[T]) CodePos() scan.TokenPos {
	return l.Pos
}
func (l Literal[T]) String() string {
	return fmt.Sprint(l.Value)
}

type NameBindingAssociation struct {
	Elem    BindingElement
	Colon   scan.TokenPos
	BoundTo NameBindingValueRef
}

func (n NameBindingAssociation) String() string {
	var bound any = n.BoundTo
	if bound == nil {
		bound = n.Elem
	}
	return fmt.Sprintf("%v: %v", n.Elem, bound)
}
func (n NameBindingAssociation) CodePos() scan.TokenPos {
	return n.Elem.CodePos()
}

type NameBinding = subBinding[NameBindingAssociation]

type InvalidBinding = subBinding[string]

type OrderBinding = subBinding[BindingElement]

type Binding struct {
	Star    scan.TokenPos
	Binding SubBinding
	Assign  scan.TokenInfo
	Value   Expr
}

func (b Binding) String() string {
	return fmt.Sprintf("*%v %v %v", b.Binding, b.Assign.RawString(), b.Value)
}
func (b Binding) CodePos() scan.TokenPos {
	return b.Star
}
