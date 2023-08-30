package parser

import (
	"errors"
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
)

type Parser struct {
	scanner *scanner.Scanner

	errors map[scanner.TokenPos]error

	astFile []ast.Ast

	output chan ast.Ast
}

type conflictResolver = func(p *Parser) func() ast.Ast

var conflictFor = map[tokens.Token]conflictResolver{
	tokens.IDENT: func(p *Parser) func() ast.Ast {
		s := p.scanner.Copy()
		for !s.CurrentToken().IsEoI() && s.CurrentToken() != tokens.EOF {
			if s.ConsumeToken().IsAssignation() {
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

func Parse(s *scanner.Scanner, conf config.ToolInfo) (chan ast.Ast, map[scanner.TokenPos]error) {
	p := Parser{output: make(chan ast.Ast)}
	p.errors = map[scanner.TokenPos]error{}
	p.scanner = s
	if conf.Mode() == config.ModeInteractive {
		go p.parseInteractive()
		return p.output, p.errors
	}

	close(p.output)
	return p.output, p.errors
}
