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

// DotExpr can be a TypeExpr too
type DotExpr struct {
	Pos    scan.TokenPos
	Idents []Ident
}

func (d *DotExpr) CodePos() scan.TokenPos {
	return d.Pos
}
func (d *DotExpr) String() string {
	str := strings.Join(array.Map(d.Idents, Ident.String), ".")
	if d.Pos != d.Idents[0].CodePos() {
		str = "." + str
	}
	return str
}
func (d *DotExpr) IsConstexpr() bool {
	for _, i := range d.Idents {
		_ = i // _, ok := i.(ConstexprDecl); if ! ok { return false }
	}
	return true
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
