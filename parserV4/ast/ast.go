package ast

import (
	"github.com/DarkMiMolle/GTL/array"
	"github.com/DarkMiMolle/GTL/optional"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"strings"
)

type Ast interface {
	CodePos() scan.TokenPos
	String() string
}

type Expr interface {
	Ast
	IsConstexpr() bool
}

type Decl interface {
	Ast
}

type TypeExpr interface {
	Expr
}

type Ident struct {
	Pos  scan.TokenPos
	Name string
	Ref  Decl
}

func (i *Ident) String() string {
	return i.Name
}
func (i *Ident) CodePos() scan.TokenPos {
	return i.Pos
}

// DotIdent can be a TypeExpr too
type DotIdent struct {
	Pos    scan.TokenPos
	Idents []Ident
}

func (d DotIdent) CodePos() scan.TokenPos {
	return d.Pos
}
func (d DotIdent) String() string {
	return strings.Join(array.MapRef(d.Idents, (*Ident).String), ".")
}

type ImportedPkg struct {
	Package DotIdent
	Renamed optional.Value[Ident]
}

func (ip ImportedPkg) String() string {
	str := ip.Package.String()

	if renamed, hasValue := ip.Renamed.LookupValue(); hasValue {
		str += " as " + renamed.String()
	}

	return str
}

type Import struct {
	Kw       scan.TokenPos
	Project  optional.Value[Literal[string]]
	Packages []ImportedPkg
}

func (i Import) CodePos() scan.TokenPos {
	return i.Kw
}

func (i Import) String() string {
	str := "import " + i.Project.String() + " "
	if len(i.Packages) == 1 {
		str += i.Packages[0].String()
		return str
	}

	str += "(\n\t" + strings.Join(array.Map(i.Packages, ImportedPkg.String), "\n\t") + ")"
	return str
}
