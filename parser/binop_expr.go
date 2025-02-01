package parser

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
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

func isBinop(t tokens.Token) bool {
	_, ok := binopPriorities[t]
	return ok
}

type binop struct {
	expr ParserOf[ast.Expr]
}

func fixeBinop(root *ast.BinopExpr) ast.BinopExpr {
	prevs := []*ast.BinopExpr{
		root,
		root,
	}

	lastPrev := func() *ast.BinopExpr {
		if len(prevs) == 0 {
			return nil
		}
		return prevs[len(prevs)-1]
	}

	current := root

	for {
		for {
			right, ok := current.Right.(*ast.BinopExpr)
			if !ok {
				break
			}

			prevs = append(prevs, right)
			current = right
		}

		for {
			left, ok := current.Left.(*ast.BinopExpr)
			if !ok {
				break
			}

			prevs = append(prevs, left)
			current = left
		}

		if _, ok := current.Right.(*ast.BinopExpr); ok {
			continue
		}

		if current == root {
			return *root
		}

		if current == lastPrev().Right {
			lastPrev().Right = *current
		} else if current == lastPrev().Left {
			lastPrev().Left = *current
		}

		current = lastPrev()
		prevs = prevs[:len(prevs)-1]

		if current == root {
			prevs = append(prevs, root)
		}
	}

	return *root
}

func organizeBinaryOperator(root *ast.BinopExpr) ast.BinopExpr {
	var (
		froms   = make(map[*ast.BinopExpr]*ast.BinopExpr)
		current = root
	)

	for {
		next, ok := current.Left.(*ast.BinopExpr)
		if !ok {
			return fixeBinop(root) // todo
		}

		froms[next] = current
		var (
			from = next
			to   = current
		)

		for up(from, to) {
			from = to
			to, ok = froms[from]

			if !ok {
				break
			}
		}

		current = from
	}
}

func up(from, to *ast.BinopExpr) bool {
	if binopPriorities[from.Op.BinaryOpToken()] < binopPriorities[to.Op.BinaryOpToken()] {
		return false
	}

	prev := *to

	to.Op = from.Op
	to.Left = from.Left
	to.Right = &prev

	prev.Left = from.Right

	return true
}

func (b binop) ContinueParsing(from ast.Expr, s scan.Scanner, errors *Errors) ast.BinopExpr {
	operator, ok := ast.GetBinopOperator(s.ConsumeToken())

	assert(ok)

	binop := &ast.BinopExpr{Left: from, Op: operator, Right: b.expr.Parse(s, errors)}

	ignoreOnce(s, tokens.NL)

	operator, ok = ast.GetBinopOperator(s.CurrentToken())

	for ok {
		s.ConsumeTokenInfo()

		binop = &ast.BinopExpr{
			Left:  binop,
			Op:    operator,
			Right: b.expr.Parse(s, errors),
		}

		operator, ok = ast.GetBinopOperator(s.CurrentToken())

		ignoreOnce(s, tokens.NL)
	}

	return organizeBinaryOperator(binop)
}
