package parser

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
)

func (p *Parser) parseLStructType(opening scanner.TokenInfo) ast.LStructType {
	lstruct := ast.LStructType{}
	lstruct.Opening = opening.FromPos()
	hasErr := false
	if hasErr = p.scanner.CurrentToken() != tokens.OBRAC; hasErr {
		p.errors[p.scanner.CurrentPos()] = fmt.Errorf("expected `{` to start the lambda structure")
		// we are trying to parse it anyway. It may be just a forgot
	}
	for p.scanner.CurrentToken() != tokens.CBRAC && p.scanner.CurrentToken() != tokens.EOF {
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
		lstruct.Attributes = append(lstruct.Attributes, attributes...)
		for range attributes {
			lstruct.Getter = append(lstruct.Getter, getter)
		}
		if p.scanner.CurrentToken() == tokens.COLON {
			p.scanner.ConsumeToken()
			continue
		}
		for p.scanner.CurrentToken().IsEoI() {
			p.scanner.ConsumeToken()
		}
	}
	if p.scanner.CurrentToken() != tokens.CBRAC {
		p.errors[p.scanner.CurrentPos()] = fmt.Errorf("missing `}` to close the structure")
	}
	return lstruct
}

func (p *Parser) parseType() ast.Ast {
	switch p.scanner.CurrentToken() {
	case tokens.IDENT:
		var typ ast.Ast = ast.Ident(p.scanner.ConsumeTokenInfo())
		var dot *ast.DottedExpr
		for p.scanner.CurrentToken() == tokens.DOT {
			dot = p.parseDotExpr(typ, p.scanner.ConsumeToken()).(*ast.DottedExpr)
			if dot.RawString {
				p.errors[dot.Right.Info().FromPos()] = fmt.Errorf("type can't have raw string dot. Maybe you wanted to surround the dotted element with 'typeof()'")
				dot.Right.Value = "/* Error here > */" + dot.Right.Value
			}
			typ = &ast.DottedType{DottedExpr: *dot}
		}
		return typ
	case tokens.TYPEOF:
	case tokens.OBRAC, tokens.STRUCT:

	case tokens.OBRAK:
	case tokens.OPAREN:

	case tokens.INTERFACE:
	case tokens.ENUM:
	case tokens.FUNC:
	}
	return nil
}
