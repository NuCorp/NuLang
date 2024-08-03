package ast

import "github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"

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
