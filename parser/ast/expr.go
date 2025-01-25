package ast

import "github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"

type Expr interface {
	ExprID() string
}

func (d DotIdent) ExprID() string {
	return "expr:ident"
}

type IntExpr int

func (i IntExpr) ExprID() string {
	return "expr:int"
}

type AskOperator struct {
	Left Expr
}

func (o AskOperator) ExprID() string {
	return "expr:.?"
}

type ForceOperator struct {
	Left Expr
}

func (o ForceOperator) ExprID() string {
	return "expr:.!"
}

var currentLv = 0

func nextLv() int {
	defer func() {
		currentLv++
	}()

	return currentLv
}

type operator struct {
	level int
	token tokens.Token
}

type BinaryOperator interface {
	BinopLevel() int
}

type binaryOperator struct {
	operator
}

func (b binaryOperator) BinopLevel() int {
	return b.level
}

var binaryOperators = map[tokens.Token]operator{
	tokens.ASKOR: {
		level: currentLv,
		token: tokens.ASKOR,
	},
}

func GetBinopOperator(t tokens.Token) (BinaryOperator, bool) {
	op, exists := binaryOperators[t]

	if !exists {
		return nil, exists
	}

	return binaryOperator{operator: op}, exists
}

type BinopExpr struct {
	Left  Expr
	Op    BinaryOperator
	Right Expr
}

type AskOrOperator struct {
	Left  Expr
	Right Expr
}

func (o AskOrOperator) ExprID() string {
	return "expr:.??."
}
