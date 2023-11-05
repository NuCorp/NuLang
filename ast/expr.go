package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"strings"
)

type Ast interface {
	From() scan.TokenPos
	To() scan.TokenPos
	String() string
}

type LiteralExpr[T comparable] struct {
	Value[T]
}

func (l LiteralExpr[T]) From() scan.TokenPos {
	return l.from.FromPos()
}
func (l LiteralExpr[T]) To() scan.TokenPos {
	return l.from.ToPos()
}
func (l LiteralExpr[T]) String() string {
	return fmt.Sprint(l.Value.Value)
}
func MakeLiteralExpr[T comparable](tokenInfo scan.TokenInfo) LiteralExpr[T] {
	return LiteralExpr[T]{Value: MakeValue[T](tokenInfo)}
}

type BinOpExpr struct {
	Left, Right Ast
	Operator    tokens.Token
	Priority    int
}

func (b *BinOpExpr) From() scan.TokenPos {
	return b.Left.From()
}
func (b *BinOpExpr) To() scan.TokenPos {
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

type IsExpr struct {
	Expr Ast
	Is   tokens.Token
	Type Type
}

func (i IsExpr) From() scan.TokenPos {
	return i.Expr.From()
}
func (i IsExpr) To() scan.TokenPos {
	return i.Type.To()
}
func (i IsExpr) String() string {
	return fmt.Sprintf("%v is %v", i.Expr, i.Type)
}

type UnOpExpr struct {
	Operator scan.TokenInfo
	Expr     Ast
}

func (u *UnOpExpr) From() scan.TokenPos {
	return u.Expr.From()
}
func (u *UnOpExpr) To() scan.TokenPos {
	return u.Operator.ToPos()
}
func (u *UnOpExpr) String() string {
	return fmt.Sprintf("%v%v", u.Expr, u.Operator.Token())
}

type SingedValue struct {
	Minus scan.TokenPos
	Value Ast
}

func (s *SingedValue) From() scan.TokenPos {
	return s.Minus
}
func (s *SingedValue) To() scan.TokenPos {
	return s.Value.To()
}
func (s *SingedValue) String() string {
	return fmt.Sprintf("-%v", s.Value)
}

type Ident scan.TokenInfo

func (s Ident) tokenInfo() scan.TokenInfo {
	return scan.TokenInfo(s)
}
func (s Ident) From() scan.TokenPos {
	return s.tokenInfo().FromPos()
}
func (s Ident) To() scan.TokenPos {
	return s.tokenInfo().ToPos()
}
func (s Ident) String() string {
	return s.tokenInfo().RawString()
}

type DottedExpr struct {
	Left      Ast
	Dot       tokens.Token
	Right     Value[string] // IDENT or STR (constexpr STR)
	RawString bool
}

func (d *DottedExpr) From() scan.TokenPos {
	return d.Left.From()
}
func (d *DottedExpr) To() scan.TokenPos {
	return d.Left.To()
}
func (d *DottedExpr) String() string {
	right := "◽"
	if d.Dot != tokens.NoInit {
		right = d.Right.Value
	}
	if d.RawString {
		right = "\"" + right + "\""
	}
	left := fmt.Sprint(d.Left)
	if d.Left == nil {
		left = ""
	}
	return fmt.Sprintf("%v.%v", left, right)
}

type AsExpr struct {
	Expr Ast

	As        tokens.Token
	Specifier tokens.Token // either !, ? or NoInit

	Type Ast
}

func (a AsExpr) From() scan.TokenPos {
	return a.Expr.From()
}
func (a AsExpr) To() scan.TokenPos {
	return a.Type.To()
}
func (a AsExpr) String() string {
	specifier := ""
	if a.Specifier != tokens.NoInit {
		specifier = a.Specifier.String()
	}
	return fmt.Sprintf("%v as%v %v", a.Expr, specifier, a.Type)
}

type TupleExpr struct {
	OpenParen  scan.TokenPos
	ExprList   []Ast
	CloseParen scan.TokenPos
}

func (t TupleExpr) From() scan.TokenPos {
	return t.OpenParen
}
func (t TupleExpr) To() scan.TokenPos {
	return t.CloseParen
}
func (t TupleExpr) String() string {
	str := "("
	for _, expr := range t.ExprList {
		str += fmt.Sprintf("%v, ", expr)
	}

	return strings.TrimSuffix(str, ", ") + ")"
}

type AnonymousStructExpr struct {
	Opening scan.TokenPos
	Fields  []BindToName
	Closing scan.TokenPos
}

func (a AnonymousStructExpr) From() scan.TokenPos {
	return a.Opening
}
func (a AnonymousStructExpr) To() scan.TokenPos {
	return a.Closing
}
func (a AnonymousStructExpr) String() string {
	str := "struct{"
	for _, field := range a.Fields {
		str += field.String() + ", "
	}
	return strings.TrimSuffix(str, ", ") + "}"
}
