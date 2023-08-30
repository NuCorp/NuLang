package parser

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
)

func (p *Parser) parseAnonymousStructType(opening scanner.TokenInfo) ast.Ast {
	lstruct := ast.AnonymousStructType{}
	lstruct.Opening = opening.FromPos()
	hasErr := false
	if hasErr = p.scanner.CurrentToken() != tokens.OBRAC; hasErr {
		p.errors[p.scanner.CurrentPos()] = fmt.Errorf("expected `{` to start the lambda structure")
		// we are trying to parse it anyway. It may be just a forgot
	} else {
		p.scanner.ConsumeToken()
	}
	for p.scanner.CurrentToken() != tokens.CBRAC && p.scanner.CurrentToken() != tokens.EOF {
		p.skipTokens(tokens.EoI()...)
		getter := p.scanner.CurrentToken() == tokens.GET
		if p.scanner.CurrentToken() == tokens.GET {
			p.scanner.ConsumeToken()
		}
		attributes := p.parseSimpleVars()
		if attributes == nil { // TODO: ERROR: check if it is ast.Error ?
			if hasErr {
				break
			}
		}
		for _, attribute := range attributes {
			lstruct.Attributes = append(lstruct.Attributes, attribute.(*ast.NamedDef))
		}
		for range attributes {
			lstruct.Getter = append(lstruct.Getter, getter)
		}
		if p.scanner.CurrentToken() == tokens.COMA {
			p.scanner.ConsumeToken()
			continue
		}
		p.skipTokens(tokens.EoI()...)
	}

	if p.scanner.CurrentToken() != tokens.CBRAC {
		p.errors[p.scanner.CurrentPos()] = fmt.Errorf("missing `}` to close the structure")
		lstruct.Ending = p.scanner.CurrentPos()
		p.skipTo(tokens.EoI()...)
		return nil // TODO: ERROR: return ast.Error ?
	}
	lstruct.Ending = p.scanner.ConsumeTokenInfo().ToPos()
	if opening.Token() == tokens.STRUCT {
		return lstruct
	}
	if p.scanner.CurrentToken() != tokens.CBRAC {
		p.errors[p.scanner.CurrentPos()] = fmt.Errorf("missing `}` to close the `{{` structure opening")
	} else {
		lstruct.Ending = p.scanner.ConsumeTokenInfo().ToPos()
	}
	return lstruct
}

func (p *Parser) parseType() ast.Ast {
	switch p.scanner.CurrentToken() {
	case tokens.IDENT:
		var typ ast.Ast = ast.Ident(p.scanner.ConsumeTokenInfo())
		var dot *ast.DottedExpr
		for p.scanner.CurrentToken() == tokens.DOT {
			dot = p.parseDotExpr(typ, p.scanner.ConsumeToken())
			if dot.RawString {
				p.errors[dot.Right.Info().FromPos()] = fmt.Errorf("type can't have raw string dot. Maybe you wanted to surround the dotted element with 'typeof()'")
				dot.Right.Value = "/* Error here > */" + dot.Right.Value
			}
			typ = &ast.DottedType{DottedExpr: *dot}
		}
		return typ
	case tokens.TYPEOF:
	case tokens.OBRAC, tokens.STRUCT:
		return p.parseAnonymousStructType(p.scanner.ConsumeTokenInfo())
	case tokens.OBRAK:
	case tokens.OPAREN:

	case tokens.INTERFACE:
	case tokens.ENUM:
	case tokens.FUNC:
	}
	return nil
}

func (p *Parser) canStartType() bool {
	return container.Contains([]tokens.Token{
		tokens.IDENT,
		tokens.TYPEOF,
		tokens.OBRAC, tokens.STRUCT,
		tokens.OBRAK, tokens.OPAREN,
		tokens.INTERFACE, tokens.ENUM, tokens.FUNC,
	}, p.scanner.CurrentToken())
}
