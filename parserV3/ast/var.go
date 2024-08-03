package ast

import (
	"github.com/DarkMiMolle/GTL/optional"
)

type Ident = string

type Var struct {
	Name  Ident
	Type  any
	Value optional.Value[any]
}
