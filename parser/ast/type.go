package ast

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
)

type Type interface {
	TypeID() string
}

type StructType struct {
	Fields       map[string]Type
	GetFields    container.Set[string]
	DefaultValue map[string]Expr
}

func (s StructType) TypeID() string {
	return "type:struct"
}

type NamedType DotIdent

func (n NamedType) TypeID() string {
	return "type:named"
}
