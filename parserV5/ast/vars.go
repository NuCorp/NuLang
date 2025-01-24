package ast

import "github.com/LicorneSharing/GTL/optional"

type Var struct {
	Name  string
	Type  Type
	Value Expr
}

type BindingAssign struct {
	NameBinding  optional.Value[NameBindingAssign]
	OrderBinding optional.Value[OrderBindingAssign]
	Value        Expr
}

func (b BindingAssign) ToVars() ([]Var, error) {
	return nil, nil
}

type NameBindingAssign struct {
	Elems       []SubBinding
	ToName      map[int]*DotIdent
	AskOrValues map[*DotIdent]AskOrOperator
	AskValues   map[*DotIdent]AskOperator
	ForceValues map[*DotIdent]ForceOperator
}

func EmptyNameBindingAssign() NameBindingAssign {
	return NameBindingAssign{
		ToName:      make(map[int]*DotIdent),
		AskOrValues: make(map[*DotIdent]AskOrOperator),
		AskValues:   make(map[*DotIdent]AskOperator),
		ForceValues: make(map[*DotIdent]ForceOperator),
	}
}

func (n NameBindingAssign) subbinding() {}

type OrderBindingAssign struct {
	Elems []SubBinding
}

func (o OrderBindingAssign) subbinding() {}

type SubBinding interface {
	subbinding()
}

func (d DotIdent) subbinding() {}
