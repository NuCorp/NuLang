package parser

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func (p *Parser) parseSimpleVars() (varElemList []ast.VarElem) {
	var vars []ast.VarElem
	for p.scanner.CurrentToken() != tokens.EOF {
		for p.scanner.CurrentTokenInfo().RawString() == "\n" {
			p.scanner.ConsumeToken()
		}
		if p.scanner.CurrentToken() != tokens.IDENT {
			p.errors[p.scanner.ConsumeTokenInfo().FromPos()] = fmt.Errorf("unexpected token")
			p.skipTo(append(tokens.EoI(), tokens.COMA, tokens.EOF)...)
			return nil // TODO: ERROR: return an AstError ?
		}
		varElem := &ast.NamedDef{}
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
				return nil // TODO: ERROR: return an AstError ?
			}
			varElem.Assign = p.scanner.ConsumeToken()
			varElem.Value = p.parseExpr()
			return []ast.VarElem{varElem}
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
			vars[i].(*ast.NamedDef).Typ = varElem.Typ
		}
		return append(vars, varElem)
	}
	panic("unreachable")
}

func (p *Parser) parseVars(kw scan.TokenInfo) ast.Ast {
	vars := &ast.VarList{Keyword: kw.FromPos()}
varElemLoop:
	for !p.scanner.CurrentToken().IsEoI() && p.scanner.CurrentToken() != tokens.EOF {
		switch p.scanner.CurrentToken() {
		case tokens.TIME: // NameBindingElem and OrderBinding
			vars.AddVars(p.parseBinding(p.scanner.ConsumeTokenInfo(), true))
		case tokens.IDENT:
			vars.AddVars(p.parseSimpleVars()...)
		default:
			p.addError(fmt.Errorf("unexpected token `%v` in variables definition", p.scanner.CurrentToken()))
		}
		if p.scanner.CurrentToken() != tokens.COMA {
			break varElemLoop
		}
		p.scanner.ConsumeToken()
		p.skipTokens(tokens.NL)
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
