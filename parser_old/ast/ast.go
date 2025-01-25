package ast

import (
	"strings"

	"github.com/LicorneSharing/GTL/optional"
	"github.com/LicorneSharing/GTL/slices"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
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
	return strings.Join(slices.MapRef(d.Idents, (*Ident).String), ".")
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

	str += "(\n\t" + strings.Join(slices.Map(i.Packages, ImportedPkg.String), "\n\t") + ")"
	return str
}

type Package struct {
	Kw   scan.TokenPos
	Name DotIdent
}

func (p Package) CodePos() scan.TokenPos {
	return p.Kw
}

func (p Package) String() string {
	return "package " + p.Name.String()
}

type File struct {
	Package Package
	Imports []Import

	Decl []Decl
}
