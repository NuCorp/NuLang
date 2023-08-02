package parser

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
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

func (p *Parser) parseSimpleVars() []ast.VarDef {
	var vars []ast.VarDef
	for p.scanner.CurrentToken() != tokens.EOF {
		for p.scanner.CurrentTokenInfo().RawString() == "\n" {
			p.scanner.ConsumeToken()
		}
		if p.scanner.CurrentToken() != tokens.IDENT {
			p.errors[p.scanner.ConsumeTokenInfo().FromPos()] = fmt.Errorf("unexpected token")
			p.skipTo(append(tokens.EoI(), tokens.COMA, tokens.EOF)...)
			return nil // TODO: return an AstError ?
		}
		varElem := ast.VarDef{}
		varElem.Name = ast.MakeValue[string](p.scanner.ConsumeTokenInfo())
		if p.scanner.CurrentToken() == tokens.COMA { // IDENT `,` IDENT --> multiple variable of the same type (without value)
			p.scanner.ConsumeToken()
			vars = append(vars, varElem)
			continue
		}
	assignment:
		if p.scanner.CurrentToken() == tokens.ASSIGN { // IDENT `=` Expr --> assignment declaration (it also might be: IDENT Type `=` Expr)
			if len(vars) != 0 {
				p.errors[p.scanner.CurrentPos()] = fmt.Errorf("unexpected '='. Cannot assign multiple variable with 1 '=', may be you wanted to use order binding")
				p.skipTo(append(tokens.EoI(), tokens.COMA, tokens.EOF)...)
				return nil
			}
			varElem.Assign = p.scanner.ConsumeToken()
			varElem.Value = p.parseExpr()
			return []ast.VarDef{varElem}
		}

		// IDENT ?? --> the '??' corresponds to the type
		// It may be IDENT Type `=` Expr (meaning it has only one varElem) or IDENT Type `,` (and then 'Type' apply to all previous varElem)
		// then return
		varElem.Typ = p.parseType()
		if p.scanner.CurrentToken() == tokens.ASSIGN {
			goto assignment // first case: IDENT Type `=` Expr --> Type is only for the one VarElem
		}
		// second case: IDENT (`,` IDENT)* Type --> Type apply to all previous VarElem
		for i := range vars {
			vars[i].Typ = varElem.Typ
		}
		return vars
	}
	panic("unreachable")
}

type nameBindingElem struct {
	nameBinding map[ast.Value[string]]ast.Value[string]
	left        ast.Ast      // left = struct{c int}{s...} // TODO: ast.Type --> deduced
	assignToken tokens.Token // either '=' or ':='
	value       ast.Ast      // expr
}

func (p *Parser) parseNameBindingAssignment(obrac scanner.TokenPos, expectedAssignment tokens.Token) nameBindingElem {
	element := nameBindingElem{nameBinding: map[ast.Value[string]]ast.Value[string]{}}
	level := 0
	for p.scanner.CurrentToken() != tokens.CBRAC || level > 0 {
	}
	return element
}

func (p *Parser) parseNameBindingVar(obrac scanner.TokenPos) []ast.VarDef {

	return nil
}

func (p *Parser) parseVars(kw scanner.TokenInfo) ast.Ast {
	vars := &ast.VarList{Keyword: kw.FromPos()}
varElemLoop:
	for !p.scanner.CurrentToken().IsEoI() && p.scanner.CurrentToken() != tokens.EOF {
		switch p.scanner.CurrentToken() {
		case tokens.OBRAK:
		case tokens.OBRAC:
		case tokens.IDENT:
			vars.AddVars(p.parseSimpleVars()...)
			if p.scanner.CurrentToken() != tokens.COMA {
				break varElemLoop
			}
			p.scanner.ConsumeToken()
			if p.scanner.CurrentToken() == tokens.NL {
				p.scanner.ConsumeToken()
			}
		}
	}
	return vars
}

func (p *Parser) parseDef() ast.Ast {
	switch p.scanner.CurrentToken() {
	case tokens.VAR:
		return p.parseVars(p.scanner.ConsumeTokenInfo())
	}
	return nil
}

func (p *Parser) parseType() ast.Ast {
	switch p.scanner.CurrentToken() {
	case tokens.IDENT:
		var typ ast.Ast = ast.Ident(p.scanner.ConsumeTokenInfo())
		var dot *ast.DottedExpr
		for p.scanner.CurrentToken() == tokens.DOT {
			dot = p.parseDotExpr(typ, p.scanner.ConsumeToken()).(*ast.DottedExpr)
			typ = dot
		}
		if dot != nil && dot.RawString {
			p.errors[dot.Right.Info().FromPos()] = fmt.Errorf("type can't have raw string dot. Maybe you wanted to surround the dotted element with 'typeof()'")
		}
		return typ
	case tokens.TYPEOF:
	case tokens.OBRAC:
	case tokens.OBRAK:
	case tokens.OPAREN:
	case tokens.STRUCT:
	case tokens.INTERFACE:
	case tokens.ENUM:
	case tokens.FUNC:
	}
	return nil
}

func (p *Parser) parseInteractive() {
	for p.scanner.CurrentToken() != tokens.EOF {
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
		for p.scanner.CurrentToken().IsEoI() {
			p.scanner.ConsumeToken()
		}
	}
}

func (p *Parser) skipTo(tokenOpt ...tokens.Token) {
	for !container.Contains(p.scanner.ConsumeToken(), tokenOpt) {
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
