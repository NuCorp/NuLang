package parserV2

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV2/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func (p *Parser) ParseConstDeclaration(constKw scan.TokenInfo) *ast.ConstDeclaration {
	scanner := p.scanner
	constDecl := &ast.ConstDeclaration{ConstKeyword: constKw}
	if scanner.CurrentToken() == tokens.PLUS {
		scanner.ConsumeTokenInfo()
		constDecl.Constexpr = true
	}
	var prevElem ast.ConstElem
	for !scanner.IsEnded() {
		switch scanner.CurrentToken() {
		case tokens.STAR: // binding
			if singleVar, ok := prevElem.(ast.SingleVarDeclaration); ok {
				if !singleVar.Type.HasValue() && !singleVar.Value.HasValue() {
					// ERROR: missing either type or value to previous variable
				}
			}
			prevElem = p.parseVarAssignNameBinding(scanner.ConsumeTokenInfo())
			constDecl.Constants = append(constDecl.Constants, prevElem)
		case tokens.IDENT: // simple
			newVar := p.parseVarSimpleElement(scanner.ConsumeTokenInfo())
			if newVar.Value.HasValue() {
				if singleVar, ok := prevElem.(ast.SingleVarDeclaration); ok {
					if !singleVar.Type.HasValue() && !singleVar.Value.HasValue() {
						// ERROR: previous variable has neither a type and a value
					}
				}
			} else if newVar.Type.HasValue() {
				for i := len(constDecl.Constants) - 1; i >= 0; i-- {
					elem := &constDecl.Constants[i]

					if singleVar, ok := (*elem).(ast.SingleVarDeclaration); ok {
						if singleVar.Type.HasValue() || singleVar.Value.HasValue() {
							break
						}
						singleVar.Type.Set(newVar.Type.Value())
						*elem = singleVar
						continue
					}
					break
				}
			}
			prevElem = newVar
			constDecl.Constants = append(constDecl.Constants, newVar)
		default:
			// error
			p.SkipTo(append(tokens.EoI(), tokens.COMA)...)
		}
		if scanner.CurrentToken() != tokens.COMA {
			break
		}
		p.Skip(tokens.NL, tokens.COMA)
	}

	return constDecl
}
