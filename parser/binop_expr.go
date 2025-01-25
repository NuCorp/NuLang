package parser

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils/maps"
)

var currentPrio = 0

func nextPrio() int {
	defer func() {
		currentPrio++
	}()

	return currentPrio
}

func samePrio() int {
	return currentPrio
}

var (
	binopPriorities = map[tokens.Token]int{
		tokens.ASKOR: currentPrio,

		tokens.TIME:     nextPrio(),
		tokens.DIV:      samePrio(),
		tokens.FRAC_DIV: samePrio(),
		tokens.MOD:      samePrio(),

		tokens.PLUS:  nextPrio(),
		tokens.MINUS: samePrio(),

		tokens.AND: nextPrio(),

		tokens.OR: nextPrio(),

		tokens.EQ:  nextPrio(),
		tokens.NEQ: samePrio(),
		tokens.GT:  samePrio(),
		tokens.LT:  samePrio(),
		tokens.GE:  samePrio(),
		tokens.LE:  samePrio(),
	}

	binaryOperators = maps.Keys(binopPriorities)
)

func organizeBinaryOperator(root *ast.BinopExpr) *ast.BinopExpr {
	current := root

	for {
		next, ok := current.Right.(*ast.BinopExpr)
		if !ok {
			return root
		}

		opCur := current.Op.BinaryOpToken()
		opNext := next.Op.BinaryOpToken()

		if binopPriorities[opCur] < binopPriorities[opNext] {
			prevNext := *next
			*next = ast.BinopExpr{
				Left:  current.Left,
				Op:    current.Op,
				Right: next.Left,
			}
			*current = ast.BinopExpr{ // current
				Left:  next, // will become newLeft
				Op:    prevNext.Op,
				Right: prevNext.Right,
			}

			/*
						OP1 (current)
						|		\
						a	 	OP2 (next)
								|	\
								b	...
				==>
						OP2 (current)
						|			\
						OP1 (next)	...
						|	\
						a	b
			*/
		} else {
			current = next
		}
	}
}

type binop struct {
	expr ParserOf[ast.Expr]
}
