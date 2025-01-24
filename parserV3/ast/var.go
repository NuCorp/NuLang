package ast

import (
	"github.com/LicorneSharing/GTL/optional"
)

type Ident = string

type Var struct {
	Name  Ident
	Type  any
	Value optional.Value[any]
}
