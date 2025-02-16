package ast

import (
	"strings"

	"github.com/LicorneSharing/GTL/optional"
	"github.com/LicorneSharing/GTL/slices"

	"github.com/NuCorp/NuLang/scan"
)

type Scope struct {
	Body []Ast
}

type Parameter struct {
	Named   optional.Value[scan.TokenPos]
	Name    Ident
	Type    TypeExpr
	Default Expr // may be nil
}

func (p *Parameter) CodePos() scan.TokenPos {
	if named, hasValue := p.Named.LookupValue(); hasValue {
		return named
	}
	return p.Name.CodePos()
}

func (p *Parameter) String() string {
	str := p.Name.String()

	if p.Named.HasValue() {
		str = "*" + str
	}

	str += p.Type.String()

	if p.Default != nil {
		str += " = " + p.Default.String()
	}

	return str
}

type FuncDecl struct {
	Func        scan.TokenPos
	IsConstexpr bool

	Name     Ident
	MayCrash bool
	MayThrow bool

	Param    []Parameter
	Variadic optional.Value[*Parameter]

	ReturnType TypeExpr

	Body Scope
}

func (f *FuncDecl) CodePos() scan.TokenPos {
	return f.Func
}

func (f *FuncDecl) String() string {
	str := "func " + f.Name.String() + "(" + strings.Join(slices.MapRef(f.Param, (*Parameter).String), ", ")

	if f.Variadic.HasValue() && len(f.Param) > 0 {
		str += ", " + f.Variadic.Get().String()
	}

	return str
}

type FuncType struct {
	Params     []Parameter
	ReturnType TypeExpr
}
