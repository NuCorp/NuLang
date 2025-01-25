package ast

import "github.com/LicorneSharing/GTL/optional"

type Package struct {
	Name DotIdent
	Defs []Def
}

type Import struct {
	Access  optional.Value[string]
	Package DotIdent
	As      optional.Value[string]
}
