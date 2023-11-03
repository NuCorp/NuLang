package parser

import (
	"errors"
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"strings"
)

type Parser struct {
	scanner scan.Scanner

	errors map[scan.TokenPos]error

	astFile []ast.Ast

	output chan ast.Ast
}

type conflictResolver = func(p *Parser) func() ast.Ast

var conflictFor = map[tokens.Token]conflictResolver{
	tokens.IDENT: func(p *Parser) func() ast.Ast {
		nextToken := func(i int) tokens.Token { return p.scanner.Next(i).Token() }
		for i := 0; !nextToken(i).IsEoI() && nextToken(i) != tokens.EOF; i++ {
			if nextToken(i).IsAssignation() {
				return func() ast.Ast {
					p.skipTo(tokens.EOF)
					p.errors[p.scanner.CurrentPos()] = fmt.Errorf("conflict not handle yet")
					return nil
				} // TODO
			}
		}
		return p.parseExpr
	},
}

func (p *Parser) addError(err error) {
	oldErr, exists := p.errors[p.scanner.CurrentPos()]
	if exists && !strings.HasPrefix(err.Error(), "|\t") {
		err = fmt.Errorf("|\t%v", err.Error())
	}
	p.errors[p.scanner.CurrentPos()] = errors.Join(oldErr, err)
}

func (p *Parser) parseInteractive() {
	for p.scanner.CurrentToken() != tokens.EOF {
		p.skipTokens(tokens.EoI()...)
		if resolver, found := conflictFor[p.scanner.CurrentToken()]; found {
			parser := resolver(p)
			astElem := parser()
			p.astFile = append(p.astFile, astElem)
			p.output <- astElem
			continue
		}
		if p.canStartExpr() {
			astElem := p.parseExpr()
			p.astFile = append(p.astFile, astElem)
			p.output <- astElem
			continue
		}
		if p.scanner.CurrentToken() == tokens.VAR {
			astElem := p.parseDef()
			p.astFile = append(p.astFile, astElem)
			p.output <- astElem
			continue
		}
		if p.scanner.CurrentToken() == tokens.EOF {
			break
		}
		p.addError(fmt.Errorf("invalid token `%v` to start an interactive instruction", p.scanner.ConsumeToken()))
		p.skipTo(tokens.EoI()...)
	}
	close(p.output)
}

func (p *Parser) skipTo(tokenOpt ...tokens.Token) {
	for p.scanner.CurrentToken() != tokens.EOF && !container.Contains(tokenOpt, p.scanner.CurrentToken()) {
		p.scanner.ConsumeTokenInfo()
	}
}

func (p *Parser) skipTokens(tokenList ...tokens.Token) {
	for container.Contains(tokenList, p.scanner.CurrentToken()) {
		p.scanner.ConsumeToken()
	}
}

func (p *Parser) parseFunctionCall(expr ast.Ast, oparent tokens.Token) ast.Ast {
	funcCall := ast.NewFunctionCall()
	funcCall.SetCaller(expr)
	funcCall.OpenParentheses(oparent)

	for p.scanner.CurrentToken() != tokens.CPAREN {
		p.skipTokens(tokens.NL)
		if p.scanner.CurrentToken() == tokens.STAR {
			funcCall.AddBoundArgument(p.parseMatchBinding(p.scanner.ConsumeTokenInfo()))
		} else if p.scanner.CurrentToken() != tokens.CPAREN {
			funcCall.AddOrderArgument(p.parseExpr())
		}
		if p.scanner.CurrentToken() == tokens.COMA {
			p.scanner.ConsumeToken()
		} else {
			p.addError(fmt.Errorf("expected `,` or `)` after function argument but got: %v", p.scanner.CurrentToken()))
			if p.scanner.CurrentToken().IsOneOf(tokens.OBRAK, tokens.OBRAC) {
				funcCall.CloseParentheses(p.scanner.ConsumeTokenInfo())
			} else if p.scanner.CurrentToken().IsOneOf(tokens.SEMI, tokens.DOT, tokens.COLON) {
				p.scanner.ConsumeToken()
				continue
			}
			break
		}
	}

	return funcCall
}

func Parse(s scan.Scanner, conf config.ToolInfo) (chan ast.Ast, map[scan.TokenPos]error) {
	p := Parser{output: make(chan ast.Ast)}
	p.errors = map[scan.TokenPos]error{}
	p.scanner = s
	if conf.Mode() == config.ModeInteractive {
		go p.parseInteractive()
		return p.output, p.errors
	}

	close(p.output)
	return p.output, p.errors
}
