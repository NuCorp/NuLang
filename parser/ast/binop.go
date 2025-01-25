package ast

import "github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"

type operator struct {
	token tokens.Token
}

type BinaryOperator interface {
	BinaryOpToken() tokens.Token
}

type binaryOperator operator

func (b binaryOperator) BinaryOpToken() tokens.Token {
	return b.token
}

var binaryOperators = map[tokens.Token]binaryOperator{
	tokens.ASKOR: {
		token: tokens.ASKOR,
	},
}

func GetBinopOperator(t tokens.Token) (BinaryOperator, bool) {
	op, exists := binaryOperators[t]

	if !exists {
		return nil, exists
	}

	return op, exists
}

type BinopExpr struct {
	Left  Expr
	Op    BinaryOperator
	Right Expr
}

func (BinopExpr) ExprID() string { return "expr:binop" }
