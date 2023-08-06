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
}

type conflictResolver = func(p *Parser) func() ast.Ast

var conflictFor = map[tokens.Token]conflictResolver{
	tokens.IDENT: func(p *Parser) func() ast.Ast {
		s := *p.scanner
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

func (p *Parser) parseNameBindingElem() ast.BindingElement {
	elem := &ast.NameBindingElem{}
	if p.scanner.CurrentToken() != tokens.IDENT {
		p.addError(fmt.Errorf("expected an identifier but got: %v", p.scanner.CurrentToken()))
		return nil // TODO: ERROR
	}
	elem.VariableName = ast.Ident(p.scanner.ConsumeTokenInfo())

	if p.scanner.CurrentToken() != tokens.COLON {
		return elem
	}

	elem.Colon = p.scanner.ConsumeToken()

	if p.scanner.CurrentToken() == tokens.ELLIPSIS {
		return &ast.BindingLeft{
			VariableName: elem.VariableName,
			Colon:        elem.Colon,
			Ellipsis:     p.scanner.ConsumeTokenInfo(),
		}
	}

	if p.scanner.CurrentToken() != tokens.IDENT {
		p.addError(fmt.Errorf("expected an identifier (corresponding to the attribute name) after `:`, but got: %v", p.scanner.CurrentToken()))
		return nil // TODO: ERROR
	}

	elem.AttributeName = ast.Ident(p.scanner.CurrentTokenInfo())

	elem.BindingExpr = p.parseExpr() // we are sure the expression starts with an identifier

	return elem
}

func (p *Parser) parseNameBinding(star, obrace scanner.TokenInfo, isVar bool) *ast.NameBinding {
	nameBinding := &ast.NameBinding{
		Star:      star.FromPos(),
		OpenBrace: obrace.Token(),
	}
	p.skipTokens(tokens.EoI()...)
	for p.scanner.CurrentToken() != tokens.CBRAC && p.scanner.CurrentToken() != tokens.EOF {
		if nameBinding.Left != nil {
			p.addError(fmt.Errorf("cannot continue binding after `...` binding"))
			p.skipTo(append(tokens.EoI(), tokens.CBRAC)...)
			break
		}
		elem := p.parseNameBindingElem()
		if elem != nil {
			switch elem := elem.(type) {
			case *ast.NameBindingElem:
				nameBinding.Elements = append(nameBinding.Elements, elem)
			case *ast.BindingLeft:
				nameBinding.Left = elem
			default:
				panic("unreachable")
			}
		}
		if p.scanner.CurrentToken() == tokens.COMA {
			p.scanner.ConsumeToken()
			p.skipTokens(tokens.EoI()...)
			continue
		}

		if p.scanner.CurrentToken() != tokens.CBRAC {
			p.skipTokens(tokens.EoI()...)
			if p.scanner.CurrentToken() != tokens.CBRAC {
				p.addError(fmt.Errorf("binding element must be directly followed by `,` to be continued"))
			}
			break
		}

	}

	if len(nameBinding.Elements) == 0 {
		p.addError(fmt.Errorf("|\tbinding element can't be empty"))
	}

	if p.scanner.CurrentToken() == tokens.CBRAC {
		nameBinding.CloseBrace = p.scanner.ConsumeToken()
	} else {
		p.addError(fmt.Errorf("|\texpected '}' to close the name binding"))
	}

	if p.scanner.CurrentToken() != tokens.ASSIGN && p.scanner.CurrentToken() != tokens.DEFINE {
		if _, exists := p.errors[p.scanner.CurrentPos()]; exists {
			p.skipTo(tokens.EoI()...)
			return nil // TODO: ERROR: return ast.Error ?
		}
		p.addError(fmt.Errorf("expected assign (=) or define (:=) token but got `%v`", p.scanner.CurrentToken()))
		p.skipTo(tokens.EoI()...)
		return nil // TODO: ERROR: return ast.Error ?
	}

	if p.scanner.CurrentToken() == tokens.DEFINE && isVar {
		p.addError(fmt.Errorf("define operator (:=) cannot be used in `var` like defininiton"))
		p.scanner.ConsumeToken()
		nameBinding.AssignToken = tokens.ASSIGN
	} else {
		nameBinding.AssignToken = p.scanner.ConsumeToken()
	}

	nameBinding.Value = p.parseExpr()

	return nameBinding
}

func (p *Parser) parseBinding(star scanner.TokenInfo, isVar bool) ast.VarElem {
	switch p.scanner.CurrentToken() {
	case tokens.OBRAC:
		return p.parseNameBinding(star, p.scanner.ConsumeTokenInfo(), isVar)
	default:
		p.errors[p.scanner.CurrentPos()] = fmt.Errorf("unexpected token %v", p.scanner.CurrentToken())
		return nil
	}
}

func (p *Parser) parseInteractive() {
	for p.scanner.CurrentToken() != tokens.EOF {
		p.skipTokens(tokens.EoI()...)
		if resolver, found := conflictFor[p.scanner.CurrentToken()]; found {
			parser := resolver(p)
			p.astFile = append(p.astFile, parser())
			continue
		}
		if p.canStartExpr() {
			p.astFile = append(p.astFile, p.parseExpr())
			continue
		}
		if p.scanner.CurrentToken() == tokens.VAR {
			p.astFile = append(p.astFile, p.parseDef())
			continue
		}
		if p.scanner.CurrentToken() == tokens.EOF {
			break
		}
		p.addError(fmt.Errorf("invalid token `%v` to start an interactive instruction", p.scanner.ConsumeToken()))
		p.skipTo(tokens.EoI()...)
	}
}

func (p *Parser) skipTo(tokenOpt ...tokens.Token) {
	for p.scanner.CurrentToken() != tokens.EOF && !container.Contains(p.scanner.ConsumeToken(), tokenOpt) {
	}
}

func (p *Parser) skipTokens(tokenList ...tokens.Token) {
	for container.Contains(p.scanner.CurrentToken(), tokenList) {
		p.scanner.ConsumeToken()
	}
}

func Parse(s scanner.Scanner, conf config.ToolInfo) ([]ast.Ast, map[scanner.TokenPos]error) {
	p := Parser{}
	p.errors = map[scanner.TokenPos]error{}
	p.scanner = &s
	if conf.Kind() == config.Interactive {
		p.parseInteractive()
		return p.astFile, p.errors
	}

	return nil, p.errors
}
