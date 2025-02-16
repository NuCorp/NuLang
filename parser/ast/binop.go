package ast

import "github.com/NuCorp/NuLang/scan/tokens"

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

func (b binaryOperator) String() string {
	return b.token.String()
}

var binaryOperators = map[tokens.Token]binaryOperator{
	tokens.ASKOR: {
		token: tokens.ASKOR,
	},
	tokens.PLUS: {
		token: tokens.PLUS,
	},
	tokens.MINUS: {
		token: tokens.MINUS,
	},
	tokens.TIME: {
		token: tokens.TIME,
	},
	tokens.DIV: {
		token: tokens.DIV,
	},
	tokens.FRAC_DIV: {
		token: tokens.FRAC_DIV,
	},
	tokens.MOD: {
		token: tokens.MOD,
	},
	tokens.EQ: {
		token: tokens.EQ,
	},
	tokens.NEQ: {
		token: tokens.NEQ,
	},
	tokens.GT: {
		token: tokens.GT,
	},
	tokens.LT: {
		token: tokens.LT,
	},
	tokens.GE: {
		token: tokens.GE,
	},
	tokens.LE: {
		token: tokens.LE,
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

func (BinopExpr) expr() {}
