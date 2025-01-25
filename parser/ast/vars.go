package ast

import (
	"slices"

	"github.com/LicorneSharing/GTL/optional"
	gtlslices "github.com/LicorneSharing/GTL/slices"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
)

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
func (n NameBindingAssign) ElemsName() []string {
	return slices.Concat(gtlslices.Map(n.Elems, SubBinding.ElemsName)...)
}

type OrderBindingAssign struct {
	Elems   []SubBinding
	Forced  container.Set[int]
	Asked   container.Set[int]
	AskedOr map[int]Expr
}

func (o OrderBindingAssign) subbinding() {}
func (o OrderBindingAssign) ElemsName() []string {
	return slices.Concat(gtlslices.Map(o.Elems, SubBinding.ElemsName)...)
}

type SubBinding interface {
	ElemsName() []string
	subbinding()
}

func (d DotIdent) subbinding() {}
func (d DotIdent) ElemsName() []string {
	return []string{d.Pack()}
}
