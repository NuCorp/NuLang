package ast

import (
	"github.com/DarkMiMolle/GTL/optional"
)

type Ast interface {
}

type Ident = string

type Var struct {
	Name   Ident
	NoInit bool
	Type   any
	Value  optional.Value[any]
}
