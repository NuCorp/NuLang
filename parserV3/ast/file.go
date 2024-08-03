package ast

import (
	"github.com/DarkMiMolle/GTL/optional"
)

type Ast interface {
}

type Package struct {
	Name Ident
}

/*

PKGs = IDENT (.IDENT)*

PRECISE = (`[` IDENT (, IDENT)* `]`)

RENAME = (as IDENT)

import IDENT? PKGs PRECISE? RENAME?

FILE = file:PATH

import `[` (FILE|git) `]` STR PKGs PRECISE? RENAME?

*/

type Import struct {
	Header   optional.Value[string] // IDENT | file:PATH | git
	Packages []Ident
	Precises []Ident
	Rename   optional.Value[Ident]
}

func (i Import) ImportName() Ident {
	name := i.Rename.ValueOr(i.Packages[len(i.Packages)-1])
	if name == "_" {
		return ""
	}
	return name
}

type File struct {
	Package Package
	Imports []Import

	Defs map[Ident]any
}
