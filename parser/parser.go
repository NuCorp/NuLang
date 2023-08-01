package parser

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils"
)

type Parser struct {
	scanner *scanner.Scanner

	errors map[scanner.TokenPos]error

	astFile []ast.Ast
}

type parsingFunction = func() ast.Ast

func (p *Parser) canStartExpr() bool {
	token := p.scanner.CurrentToken()
	if token.IsLiteral() {
		return true
	}
	switch token {
	case tokens.OBRAC, tokens.OBRAK, tokens.OPAREN,
		tokens.IDENT,
		tokens.MINUS:
		return true
	}
	return false
}

var conflictFor = map[tokens.Token]parsingFunction{}

var binaryPriority = 0

func nextPriority() int {
	binaryPriority++
	return binaryPriority
}
func samePriority() int {
	return binaryPriority
}

var priorityForBinOp = map[tokens.Token]int{
	tokens.PLUS:  samePriority(),
	tokens.MINUS: samePriority(),

	tokens.TIME: nextPriority(),
	tokens.DIV:  samePriority(),
	tokens.MOD:  samePriority(),

	tokens.FRAC_DIV: nextPriority(),
}

func (p *Parser) parseBinop(left ast.Ast, operator tokens.Token) ast.Ast {
	right := p.parseSingleExpr()
	priority, found := priorityForBinOp[operator]
	if !found {
		panic("shouldn't be here")
	}
	root, ok := left.(*ast.BinOpExpr)
	if !ok {
		return ast.MakeBinOpExpr(left, right, operator, priority)
	}

	binop := root

	for binop.Priority < priority {
		if binopRight, ok := binop.Right.(*ast.BinOpExpr); ok {
			binop = binopRight
		} else {
			newBinop := &ast.BinOpExpr{
				Left:     binop.Right,
				Right:    right,
				Operator: operator,
				Priority: priority,
			}
			binop.Right = newBinop
			return root
		}
	}
	temp_binop := *binop
	*binop = ast.BinOpExpr{
		Left:     &temp_binop,
		Right:    right,
		Operator: operator,
		Priority: priority,
	}

	return root
}

func (p *Parser) parseSingleExpr() ast.Ast {
	var expr ast.Ast
	if p.scanner.CurrentToken().IsLiteral() {
		expr = p.parseLiteralValue()
	}
	return expr
}

func (p *Parser) parseExpr() ast.Ast {
	expr := p.parseSingleExpr()
	for p.scanner.CurrentToken() != tokens.EoI && p.scanner.CurrentToken() != tokens.EOF {
		switch p.scanner.CurrentToken() {
		case tokens.PLUS, tokens.MINUS, tokens.TIME, tokens.DIV, tokens.MOD, tokens.FRAC_DIV:
			expr = p.parseBinop(expr, p.scanner.ConsumeToken())
		}
	}
	return expr
}

func (p *Parser) parseLiteralValue() ast.Ast {
	scan := p.scanner
	switch scan.CurrentToken() {
	case tokens.INT:
		return ast.MakeLiteralExpr[uint](scan.ConsumeTokenInfo())
	case tokens.STR:
		return ast.MakeLiteralExpr[string](scan.ConsumeTokenInfo())
	case tokens.FLOAT:
		return ast.MakeLiteralExpr[float64](scan.ConsumeTokenInfo())
	case tokens.FRACTION:
		return ast.MakeLiteralExpr[utils.Fraction](scan.ConsumeTokenInfo())
	case tokens.CHAR:
		return ast.MakeLiteralExpr[rune](scan.ConsumeTokenInfo())
	case tokens.TRUE, tokens.FALSE:
		return ast.MakeLiteralExpr[bool](scan.ConsumeTokenInfo())
	default:
		panic("invalid call - shouldn't be here") // unreachable
	}
}

func (p *Parser) parseInteractive() {
	for p.scanner.CurrentToken() != tokens.EOF {
		if resolver, found := conflictFor[p.scanner.CurrentToken()]; found {
			p.astFile = append(p.astFile, resolver())
			continue
		}
		if p.canStartExpr() {
			p.astFile = append(p.astFile, p.parseExpr())
		}
	}
}

func Parse(s scanner.Scanner, conf config.ToolInfo) ([]ast.Ast, map[scanner.TokenPos]error) {
	p := Parser{}
	p.scanner = &s
	if conf.Kind() == config.Interactive {
		p.parseInteractive()
		return p.astFile, p.errors
	}

	return nil, p.errors
}
