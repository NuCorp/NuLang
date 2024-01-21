package ast

import (
	"github.com/DarkMiMolle/GTL/optional"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
)

type Position = scan.TokenPos

type Keyword = scan.TokenInfo

type Ident struct {
	scan.TokenInfo
}

type DotIdent struct {
	IdentStart optional.Value[*Ident]
	DotStart   optional.Value[Position]

	Idents []Ident
}

/*
At will give the ident at the given position or nil

---

If the DotIdent.IdentStart is missing like in: `.attribute` or `.return`
then At(0) should return nil and then At(1) should return DotIdent.Idents[0]

However if DotIdent.IdentStart is set like in: `a.b`
then At(0) should return &DotIdent.Idents[0] and At(1) DotIdent.Idents[1]
*/
func (dot DotIdent) At(idx int) *Ident {
	if idx == 0 {
		return dot.IdentStart.ValueOr(nil)
	}
	if !dot.IdentStart.HasValue() {
		idx--
	}
	if idx >= len(dot.Idents) {
		return nil
	}
	return &dot.Idents[idx]
}

type Package struct {
	PkgKeyword Keyword
	Name       DotIdent
}

type Ast interface{}

type File struct {
	Name        string
	Path        string
	PackagePath []string

	Errors map[scan.TokenPos]error

	Package *Package
	Import  *Import

	Code []Ast
}
