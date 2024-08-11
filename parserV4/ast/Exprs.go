package ast

import (
	"fmt"
	"github.com/DarkMiMolle/GTL/array"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"strings"
)

type LiteralExpr struct {
	Pos   scan.TokenPos
	Value any
}

func (l LiteralExpr) CodePos() scan.TokenPos {
	return l.Pos
}
func (l LiteralExpr) String() string {
	return fmt.Sprint(l.Value)
}
func (l LiteralExpr) IsConstexpr() bool {
	return true
}

// IdentExpr can be a TypeExpr too
type IdentExpr struct {
	Ident
}

func (i *IdentExpr) IsConstexpr() bool {
	//TODO implemnt me
	panic("implement me")
}

type DotExpr struct {
	Pos    scan.TokenPos
	First  Expr // can be IdentExpr
	Idents []Ident
}

func (d *DotExpr) CodePos() scan.TokenPos {
	return d.Pos
}

func (d *DotExpr) String() string {
	str := ""
	if d.First == nil {
		str = "."
	}
	return str + strings.Join(array.MapRef(d.Idents, (*Ident).String), ".")
}

func (d *DotExpr) IsConstexpr() bool {
	//TODO implement me
	panic("implement me")
}

type Operator string

type BinaryExpr struct {
	Left  Expr
	Op    Operator
	Right Expr
}

func (b *BinaryExpr) CodePos() scan.TokenPos {
	return b.Left.CodePos()
}

func (b *BinaryExpr) String() string {
	return b.Left.String() + " " + string(b.Op) + " " + b.Right.String()
}

func (b *BinaryExpr) IsConstexpr() bool {
	//TODO implement me
	panic("implement me")
	// TODO: Expr.Type() TypeInfo
	// 	TypeInfo.Operators().Get(b.Op).IsConstexpr() <- possible only after type checking
}
