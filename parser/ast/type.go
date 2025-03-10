package ast

import (
	"github.com/NuCorp/NuLang/container"
)

type Type interface {
	AsType() Type
}

type StructType struct {
	Fields       map[string]Type
	GetFields    container.Set[string]
	DefaultValue map[string]Expr
}

func (s StructType) AsType() Type { return s }

type NamedType = DotIdent

func (n NamedType) AsType() Type { return n }

func (f FuncType) AsType() Type { return f }
