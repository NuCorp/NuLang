package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
)

type Ast interface {
	From() scanner.TokenPos
	To() scanner.TokenPos
}

type LiteralExpr[T comparable] struct {
	Value[T]
}

func (l LiteralExpr[T]) From() scanner.TokenPos {
	return l.from.FromPos()
}
func (l LiteralExpr[T]) To() scanner.TokenPos {
	return l.from.ToPos()
}
func (l LiteralExpr[T]) String() string {
	return fmt.Sprint(l.Value.Value)
}
func MakeLiteralExpr[T comparable](tokenInfo scanner.TokenInfo) LiteralExpr[T] {
	return LiteralExpr[T]{Value: MakeValue[T](tokenInfo)}
}

type BinOpExpr struct {
	Left, Right Ast
	Operator    tokens.Token
	Priority    int
}

func (b *BinOpExpr) From() scanner.TokenPos {
	return b.Left.From()
}
func (b *BinOpExpr) To() scanner.TokenPos {
	return b.Right.To()
}
func (b *BinOpExpr) String() string {
	return fmt.Sprintf("(%v %v %v)", b.Left, b.Operator, b.Right)
}
func MakeBinOpExpr(left, right Ast, operator tokens.Token, priority int) *BinOpExpr {
	return &BinOpExpr{
		Left:     left,
		Right:    right,
		Operator: operator,
		Priority: priority,
	}
}

type UnOpExpr struct {
	Operator scanner.TokenInfo
	Expr     Ast
}

func (u *UnOpExpr) From() scanner.TokenPos {
	return u.Expr.From()
}
func (u *UnOpExpr) To() scanner.TokenPos {
	return u.Operator.ToPos()
}
func (u *UnOpExpr) String() string {
	return fmt.Sprintf("%v%v", u.Expr, u.Operator.Token())
}

type SingedValue struct {
	Minus scanner.TokenPos
	Value Ast
}

func (s *SingedValue) From() scanner.TokenPos {
	return s.Minus
}
func (s *SingedValue) To() scanner.TokenPos {
	return s.Value.To()
}
func (s *SingedValue) String() string {
	return fmt.Sprintf("-%v", s.Value)
}